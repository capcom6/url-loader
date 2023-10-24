package loader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type Config struct {
	Timeout time.Duration
	Buffer  uint
	UseHEAD bool
}

// Loader loads URLs
type Loader struct {
	Client *http.Client

	Timeout time.Duration
	Buffer  uint
	UseHEAD bool
}

func New(client *http.Client, cfg Config) *Loader {
	if cfg.Buffer == 0 {
		cfg.Buffer = 32 * 1024
	}

	return &Loader{
		Client:  client,
		Timeout: cfg.Timeout,
		Buffer:  cfg.Buffer,
		UseHEAD: cfg.UseHEAD,
	}
}

// Load URL returning stats
func (l *Loader) Load(ctx context.Context, rawUrl string) (stats Stats, err error) {
	ctx, cancel := context.WithTimeout(ctx, l.Timeout)
	defer cancel()

	startTime := time.Now()
	if l.UseHEAD {
		stats, err = l.statsByHead(ctx, rawUrl)
	} else {
		stats, err = l.statsByGet(ctx, rawUrl)
	}

	// из задачи не очевидно что считать временем обработки, потому в данном случае считает все, начиная с соединения и заказнчивая загрузкой данных
	stats.Time = time.Since(startTime)

	return
}

func (l *Loader) statsByHead(ctx context.Context, rawUrl string) (Stats, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, rawUrl, nil)
	if err != nil {
		return Stats{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := l.Client.Do(req)
	if err != nil {
		return Stats{}, fmt.Errorf("failed to load url: %w", err)
	}
	defer func() {
		// необходимо вычитывать все данные из тела ответа, в противном случае возникает утечка памяти
		l.countBytes(ctx, resp.Body)
		_ = resp.Body.Close()
	}()

	contentLength := resp.Header.Get("Content-Length")
	if contentLength == "" {
		return Stats{}, fmt.Errorf("Content-Length header is missing")
	}

	length, err := strconv.ParseUint(contentLength, 10, 64)
	if err != nil {
		return Stats{}, fmt.Errorf("failed to parse Content-Length header: %w", err)
	}

	return Stats{
		Url:  rawUrl,
		Size: length,
	}, nil
}

func (l *Loader) statsByGet(ctx context.Context, rawUrl string) (Stats, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawUrl, nil)
	if err != nil {
		return Stats{}, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := l.Client.Do(req)
	if err != nil {
		return Stats{}, fmt.Errorf("failed to load url: %w", err)
	}
	defer resp.Body.Close()

	totalBytes, err := l.countBytes(ctx, resp.Body)
	if err != nil {
		return Stats{}, fmt.Errorf("failed to count bytes: %w", err)
	}

	return Stats{
		Url:  rawUrl,
		Size: totalBytes,
	}, nil
}

// Count bytes from reader
func (l *Loader) countBytes(ctx context.Context, r io.Reader) (uint64, error) {
	totalBytes := uint64(0)
	buffer := make([]byte, l.Buffer)
	for {
		select {
		case <-ctx.Done():
			return totalBytes, ctx.Err()
		default:
			// самым простым решением было бы использовать io.Copy(io.Discard, r), но он не работает с контекстом
			// в теории же может встретится сервис, выдающий бесконечный поток данных и тогда корректно прервать выполнение будет невозможно
			n, err := r.Read(buffer)
			if err != nil {
				if err == io.EOF {
					return totalBytes + uint64(n), nil
				}
				return totalBytes, fmt.Errorf("failed to read body: %w", err)
			}
			totalBytes += uint64(n)
		}
	}
}

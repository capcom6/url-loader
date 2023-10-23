package loader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Config struct {
	Timeout time.Duration
	Buffer  uint
}

// Loader loads URLs
type Loader struct {
	Client *http.Client

	Timeout time.Duration
	Buffer  uint
}

func New(client *http.Client, cfg Config) *Loader {
	if cfg.Buffer == 0 {
		cfg.Buffer = 32 * 1024
	}

	return &Loader{
		Client:  client,
		Timeout: cfg.Timeout,
		Buffer:  cfg.Buffer,
	}
}

// Load URL returning stats
func (l *Loader) Load(ctx context.Context, rawUrl string) (Stats, error) {
	ctx, cancel := context.WithTimeout(ctx, l.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawUrl, nil)
	if err != nil {
		return Stats{}, fmt.Errorf("failed to create request: %w", err)
	}

	startTime := time.Now()
	resp, err := l.Client.Do(req)
	if err != nil {
		return Stats{}, fmt.Errorf("failed to load url: %w", err)
	}
	defer resp.Body.Close()

	totalBytes, err := l.countBytes(ctx, resp.Body)
	if err != nil {
		return Stats{}, fmt.Errorf("failed to count bytes: %w", err)
	}

	// из задачи не очевидно что считать временем обработки, потому в данном случае считает все, начиная с соединения и заказнчивая загрузкой данных
	processTime := time.Since(startTime)

	return Stats{
		Url:  rawUrl,
		Time: processTime,
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

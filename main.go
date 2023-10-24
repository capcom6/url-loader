package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"

	"github.com/capcom6/url-loader/internal/config"
	"github.com/capcom6/url-loader/internal/linereader"
	"github.com/capcom6/url-loader/internal/loader"
)

func main() {
	if err := config.Parse(); err != nil {
		fmt.Printf("An error occurred: %s\n", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	wg := &sync.WaitGroup{}

	ch := runReader(ctx, wg, config.Reader)

	runLoader(ctx, wg, ch, config.Loader)

	wg.Wait()
}

func runReader(ctx context.Context, wg *sync.WaitGroup, cfg config.ReaderConfig) <-chan string {
	ch := make(chan string)

	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
			close(ch)
		}()

		for _, v := range cfg.Filenames {
			if v == "-" {
				v = os.Stdin.Name()
			}

			f, err := os.Open(v)
			if err != nil {
				log.Printf("Failed to open %s: %s\n", v, err)
				continue
			}
			defer f.Close()

			reader := linereader.New(f)
			if err := reader.Skip(cfg.Skip); err != nil {
				log.Printf("Failed to skip lines in %s: %s\n", v, err)
				break
			}
			for {
				line, err := reader.ReadLine()
				if err == io.EOF {
					break
				}
				if err != nil {
					log.Printf("Failed to read line from %s: %s\n", v, err)
					break
				}

				select {
				case <-ctx.Done():
					return
				case ch <- line:
				}
			}
		}
	}()

	return ch
}

func runLoader(ctx context.Context, wg *sync.WaitGroup, urls <-chan string, cfg config.LoaderConfig) {
	client := http.DefaultClient
	if !cfg.FollowRedirects {
		client = &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	}

	loader := loader.New(client, loader.Config{
		Timeout: cfg.Timeout,
		Buffer:  cfg.Buffer,
		UseHEAD: cfg.UseHEAD,
	})

	for i := 0; i < cfg.Parallel; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for v := range urls {
				stats, err := loader.Load(ctx, v)
				if err != nil {
					log.Printf("Failed to load %s: %s\n", v, err)
					continue
				}

				fmt.Printf("Url: %s, Size: %d, Time: %s\n", stats.Url, stats.Size, stats.Time)
			}
		}()
	}
}

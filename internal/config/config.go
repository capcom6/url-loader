package config

import (
	"flag"
	"fmt"
	"os"
	"path"
	"runtime"
	"time"
)

type ReaderConfig struct {
	Skip      uint
	Filenames []string
}

type LoaderConfig struct {
	Parallel        int
	Timeout         time.Duration
	FollowRedirects bool
	Buffer          uint
	UseHEAD         bool
}

var Reader ReaderConfig
var Loader LoaderConfig

// Checks if all required fields are set
func Parse() error {
	flag.CommandLine.SetOutput(os.Stdout)
	flag.Usage = Help
	// всегда есть риск, что обработка прервется по независимым причинам и стоит предусмотреть возможность запуска с момента прерывания
	flag.UintVar(&Reader.Skip, "reader-skip", 0, "skip N lines")

	// поскольку тут по большей части I/O, то можно запускать много больше горутин, чем ядер, а потому стоит сделать настройкой
	flag.IntVar(&Loader.Parallel, "loader-parallel", runtime.GOMAXPROCS(0), "parallel requests")
	// чтобы избежать зависания горутин на бесконечной установке соединения или бесконечном потоке данных
	flag.DurationVar(&Loader.Timeout, "loader-timeout", time.Second, "timeout")
	// из задачи не очевидно как поступать с перенаправлениями, потому сделано настройкой
	flag.BoolVar(&Loader.FollowRedirects, "loader-redirects", true, "follow redirects")
	// скорость загрузки может существенно зависеть от размера буфера, потому также настройка
	flag.UintVar(&Loader.Buffer, "loader-buffer", 32*1024, "buffer size in bytes")
	// в ТЗ явно сказано, что необходимо получить контент, но для получения его размера можно использовать и заголовок Content-Length
	flag.BoolVar(&Loader.UseHEAD, "loader-use-head", false, "use HEAD request and Content-Length header")
	flag.Parse()
	Reader.Filenames = flag.Args()

	return validate()
}

// Prints the help message
func Help() {
	fmt.Println("URL Loader")
	printVersion()
	fmt.Printf("\nUsage: %s [options] filename [filenames...]\n", path.Base(os.Args[0]))
	flag.PrintDefaults()
}

func validate() error {
	if len(Reader.Filenames) == 0 {
		return fmt.Errorf("at least one filename is required")
	}
	return nil
}

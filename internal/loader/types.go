package loader

import "time"

type Stats struct {
	Url  string
	Time time.Duration
	Size uint64
}

package linereader

import (
	"bufio"
	"io"
)

// LineReader reads input by lines
type LineReader struct {
	scanner *bufio.Scanner
}

func New(handle io.Reader) *LineReader {
	return &LineReader{
		scanner: bufio.NewScanner(handle),
	}
}

// Read next line
func (r *LineReader) ReadLine() (string, error) {
	if !r.scanner.Scan() {
		err := r.scanner.Err()
		if err == nil {
			err = io.EOF
		}
		return "", err
	}
	return r.scanner.Text(), nil
}

// Skip lines
func (r *LineReader) Skip(lines uint) error {
	for i := uint(0); i < lines; i++ {
		if _, err := r.ReadLine(); err != nil {
			return err
		}
	}
	return nil
}

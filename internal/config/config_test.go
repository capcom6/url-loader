package config

import (
	"os"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	// Prepare the command-line arguments
	args := []string{
		"url-loader",
		"--reader-skip", "10",
		"--loader-parallel", "5",
		"--loader-timeout", "10s",
		"--loader-redirects=false",
		"file1.txt",
		"file2.txt",
	}

	// Save the original flag values
	originalArgs := os.Args
	originalSkip := Reader.Skip
	originalParallel := Loader.Parallel
	originalTimeout := Loader.Timeout
	originalFollowRedirects := Loader.FollowRedirects
	originalFilenames := Reader.Filenames

	// Run the Parse function
	os.Args = args
	if err := Parse(); err != nil {
		t.Error(err)
	}

	// Restore the original flag values after the test
	defer func() {
		os.Args = originalArgs
		Reader.Skip = originalSkip
		Loader.Parallel = originalParallel
		Loader.Timeout = originalTimeout
		Loader.FollowRedirects = originalFollowRedirects
		Reader.Filenames = originalFilenames
	}()

	// Verify the parsed values
	if Reader.Skip != 10 {
		t.Errorf("unexpected value for Reader.Skip: got %d, want %d", Reader.Skip, 10)
	}
	if Loader.Parallel != 5 {
		t.Errorf("unexpected value for Loader.Parallel: got %d, want %d", Loader.Parallel, 5)
	}
	if Loader.Timeout != 10*time.Second {
		t.Errorf("unexpected value for Loader.Timeout: got %v, want %v", Loader.Timeout, 10*time.Second)
	}
	if Loader.FollowRedirects != false {
		t.Errorf("unexpected value for Loader.FollowRedirects: got %v, want %v", Loader.FollowRedirects, false)
	}
	if len(Reader.Filenames) != 2 {
		t.Errorf("unexpected number of filenames: got %d, want %d", len(Reader.Filenames), 2)
	}
	if Reader.Filenames[0] != "file1.txt" || Reader.Filenames[1] != "file2.txt" {
		t.Errorf("unexpected filenames: got %s, want %s", Reader.Filenames, []string{"file1.txt", "file2.txt"})
	}
}

func Test_validate(t *testing.T) {
	tests := []struct {
		name    string
		reader  ReaderConfig
		wantErr bool
	}{
		{
			name: "No files",
			reader: ReaderConfig{
				Filenames: []string{},
			},
			wantErr: true,
		},
		{
			name: "Single file",
			reader: ReaderConfig{
				Filenames: []string{"file1.txt"},
			},
			wantErr: false,
		},
		{
			name: "Multiple files",
			reader: ReaderConfig{
				Filenames: []string{"file1.txt", "file2.txt"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalReader := Reader
			Reader = tt.reader
			defer func() {
				Reader = originalReader
			}()

			if err := validate(); (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

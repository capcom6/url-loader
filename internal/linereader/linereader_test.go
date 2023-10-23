package linereader

import (
	"errors"
	"io"
	"testing"
)

// MockReader is a mock implementation of io.Reader for testing purposes
type MockReader struct {
	Data []byte
	Pos  int
}

func (r *MockReader) Read(p []byte) (int, error) {
	if r.Pos >= len(r.Data) {
		return 0, io.EOF
	}
	n := copy(p, r.Data[r.Pos:])
	r.Pos += n
	return n, nil
}

func TestLineReader_ReadLine(t *testing.T) {
	type fields struct {
		reader io.Reader
	}
	tests := []struct {
		name    string
		fields  fields
		want    []string
		wantErr bool
	}{
		{
			name: "Empty",
			fields: fields{
				reader: &MockReader{Data: []byte("")},
			},
			want:    []string{},
			wantErr: false,
		},
		{
			name: "Single line w/o newline",
			fields: fields{
				reader: &MockReader{Data: []byte("line 1")},
			},
			want:    []string{"line 1"},
			wantErr: false,
		},
		{
			name: "Single line w/ newline",
			fields: fields{
				reader: &MockReader{Data: []byte("line 1\n")},
			},
			want:    []string{"line 1"},
			wantErr: false,
		},
		{
			name: "Multiple lines",
			fields: fields{
				reader: &MockReader{Data: []byte("line 1\nline 2\nline 3\n")},
			},
			want:    []string{"line 1", "line 2", "line 3"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.fields.reader)

			for _, v := range tt.want {
				got, err := r.ReadLine()
				if (err != nil) != tt.wantErr {
					t.Errorf("LineReader.ReadLine() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != v {
					t.Errorf("LineReader.ReadLine() = %v, want %v", got, tt.want)
				}
			}
			got, err := r.ReadLine()
			if !errors.Is(err, io.EOF) {
				t.Errorf("LineReader.ReadLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != "" {
				t.Errorf("LineReader.ReadLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLineReader_Skip(t *testing.T) {
	type fields struct {
		reader io.Reader
	}
	type args struct {
		lines uint
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
		wantEof bool
	}{
		{
			name: "Zero skip",
			fields: fields{
				reader: &MockReader{Data: []byte("line 1\nline 2\nline 3\n")},
			},
			args: args{
				lines: 0,
			},
			want:    "line 1",
			wantErr: false,
			wantEof: false,
		},
		{
			name: "Skip 1",
			fields: fields{
				reader: &MockReader{Data: []byte("line 1\nline 2\nline 3\n")},
			},
			args: args{
				lines: 1,
			},
			want:    "line 2",
			wantErr: false,
			wantEof: false,
		},
		{
			name: "Skip all",
			fields: fields{
				reader: &MockReader{Data: []byte("line 1\nline 2\nline 3\n")},
			},
			args: args{
				lines: 3,
			},
			want:    "",
			wantErr: false,
			wantEof: true,
		},
		{
			name: "Skip error",
			fields: fields{
				reader: &MockReader{Data: []byte("line 1\nline 2\nline 3\n")},
			},
			args: args{
				lines: 4,
			},
			want:    "",
			wantErr: true,
			wantEof: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := New(tt.fields.reader)
			if err := r.Skip(tt.args.lines); (err != nil) != tt.wantErr {
				t.Errorf("LineReader.Skip() error = %v, wantErr %v", err, tt.wantErr)
			}
			got, err := r.ReadLine()
			if (err != nil) != tt.wantEof {
				t.Errorf("LineReader.ReadLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LineReader.ReadLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLineReader_ReadLine_ErrorReading(t *testing.T) {
	// Create a LineReader with a reader that always returns an error
	reader := New(&MockReader{
		Data: []byte("line 1\nline 2\nline 3\n"),
	})

	// Set the scanner to return an error
	reader.scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		return 0, nil, errors.New("error reading")
	})

	// Read the line, expecting an error
	_, err := reader.ReadLine()
	if err == nil {
		t.Error("expected an error, but got nil")
	}
	expectedError := "error reading"
	if err.Error() != expectedError {
		t.Errorf("unexpected error message: got %s, want %s", err.Error(), expectedError)
	}
}

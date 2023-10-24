package loader

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLoader_DefaultBuffer(t *testing.T) {
	loader := New(http.DefaultClient, Config{
		Timeout: 5 * time.Second,
	})

	if loader.Buffer != 32*1024 {
		t.Errorf("expected 32*1024, got %d", loader.Buffer)
	}
}

func TestLoader_LoadTable(t *testing.T) {
	type fields struct {
		Timeout   time.Duration
		Buffer    uint
		UseHeader bool
	}
	type args struct {
		ctx     context.Context
		handler func(w http.ResponseWriter, r *http.Request)
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Stats
		wantErr bool
	}{
		{
			name: "Empty server response",
			fields: fields{
				Timeout: 10 * time.Millisecond,
				Buffer:  32 * 1024,
			},
			args: args{
				ctx: context.Background(),
				handler: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				},
			},
			want: Stats{
				Time: 0,
				Size: 0,
			},
			wantErr: false,
		},
		{
			name: "Empty response with 500 status",
			fields: fields{
				Timeout: 10 * time.Millisecond,
				Buffer:  32 * 1024,
			},
			args: args{
				ctx: context.Background(),
				handler: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				},
			},
			want: Stats{
				Size: 0,
			},
			wantErr: false,
		},
		{
			name: "With body",
			fields: fields{
				Timeout: 10 * time.Millisecond,
				Buffer:  32 * 1024,
			},
			args: args{
				ctx: context.Background(),
				handler: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					io.WriteString(w, "This is the response body")
				},
			},
			want: Stats{
				Size: 25,
			},
			wantErr: false,
		},
		{
			name: "Timeout",
			fields: fields{
				Timeout: 10 * time.Millisecond,
				Buffer:  32 * 1024,
			},
			args: args{
				ctx: context.Background(),
				handler: func(w http.ResponseWriter, r *http.Request) {
					time.Sleep(100 * time.Millisecond)
					w.WriteHeader(http.StatusOK)
				},
			},
			want:    Stats{},
			wantErr: true,
		},
		{
			name: "Use head",
			fields: fields{
				Timeout:   10 * time.Millisecond,
				Buffer:    32 * 1024,
				UseHeader: true,
			},
			args: args{
				ctx: context.Background(),
				handler: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					io.WriteString(w, "This is the response body")
				},
			},
			want: Stats{
				Size: 25,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.args.handler))
			defer server.Close()

			l := New(http.DefaultClient, Config{
				Timeout: tt.fields.Timeout,
				Buffer:  tt.fields.Buffer,
				UseHEAD: tt.fields.UseHeader,
			})
			got, err := l.Load(tt.args.ctx, server.URL)
			if (err != nil) != tt.wantErr {
				t.Errorf("Loader.Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Size != tt.want.Size {
				t.Errorf("Loader.Load() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoader_Load_ErrorCreatingRequest(t *testing.T) {
	// Create a Loader with a custom client
	client := &http.Client{}
	loader := New(client, Config{
		Timeout: 5 * time.Second,
		Buffer:  32 * 1024,
	})

	// Attempt to load an invalid URL, expecting an error
	_, err := loader.Load(context.Background(), "invalid.url")
	if err == nil {
		t.Error("expected an error, but got nil")
	}
}

package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{name: "http", url: "http://example.com", want: true},
		{name: "uppercase scheme", url: "HTTP://example.com", want: true},
		{name: "https", url: "https://example.com/path", want: true},
		{name: "missing scheme", url: "example.com"},
		{name: "unsupported scheme", url: "ftp://example.com"},
		{name: "missing host", url: "https:///path"},
		{name: "malformed", url: "https://[::1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateURL(tt.url)
			if (err == nil) != tt.want {
				t.Errorf("validateURL(%q) error = %v, want valid = %t", tt.url, err, tt.want)
			}
		})
	}
}

func TestRunRejectsInvalidArguments(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "no URLs"},
		{name: "invalid timeout", args: []string{"-timeout", "0s", "http://example.com"}},
		{name: "invalid URL", args: []string{"ftp://example.com"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if code := run(tt.args, &stdout, &stderr); code != 2 {
				t.Fatalf("run() = %d, want 2; stderr = %q", code, stderr.String())
			}
		})
	}
}

func TestRunChecksURLsInParallel(t *testing.T) {
	arrived := make(chan struct{}, 2)
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		arrived <- struct{}{}
		<-release
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	var stdout, stderr bytes.Buffer
	done := make(chan int, 1)
	go func() {
		done <- run([]string{server.URL, server.URL}, &stdout, &stderr)
	}()

	for i := 0; i < 2; i++ {
		select {
		case <-arrived:
		case <-time.After(time.Second):
			close(release)
			<-done
			t.Fatal("checks did not start concurrently")
		}
	}
	close(release)

	if code := <-done; code != 0 {
		t.Fatalf("run() = %d, want 0; stdout = %q, stderr = %q", code, stdout.String(), stderr.String())
	}
	if got := strings.Count(stdout.String(), "HTTP 204"); got != 2 {
		t.Errorf("output contains %d successful results, want 2: %s", got, stdout.String())
	}
}

func TestRunFailsOnNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	var stdout, stderr bytes.Buffer
	if code := run([]string{server.URL}, &stdout, &stderr); code != 1 {
		t.Fatalf("run() = %d, want 1; stdout = %q, stderr = %q", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "HTTP 503") {
		t.Errorf("output = %q, want HTTP 503 status", stdout.String())
	}
}

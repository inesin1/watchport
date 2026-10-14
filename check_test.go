package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheckRecordsStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want %s", r.Method, http.MethodGet)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	result := check("HTTP"+server.URL[4:], &http.Client{Timeout: time.Second})
	if result.err != nil {
		t.Fatalf("check() error = %v", result.err)
	}
	if result.statusCode != http.StatusNoContent {
		t.Errorf("statusCode = %d, want %d", result.statusCode, http.StatusNoContent)
	}
	if result.elapsed <= 0 {
		t.Errorf("elapsed = %s, want positive duration", result.elapsed)
	}
}

func TestCheckTimesOut(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	result := check(server.URL, &http.Client{Timeout: 10 * time.Millisecond})
	if result.err == nil {
		t.Fatal("check() error = nil, want timeout error")
	}
}

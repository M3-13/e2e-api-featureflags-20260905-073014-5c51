package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestLoggingLogsMethodStatusAndPath(t *testing.T) {
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	defer logger.SetOutput(os.Stdout)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := Logging(next)

	req := httptest.NewRequest(http.MethodGet, "/flags/myflag", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	line := buf.String()
	if !strings.Contains(line, "GET") {
		t.Fatalf("log line missing method: %q", line)
	}
	if !strings.Contains(line, "200") {
		t.Fatalf("log line missing status 200: %q", line)
	}
	if !strings.Contains(line, "/flags/myflag") {
		t.Fatalf("log line missing path: %q", line)
	}
}

func TestLoggingLogs500(t *testing.T) {
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	defer logger.SetOutput(os.Stdout)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	h := Logging(next)

	req := httptest.NewRequest(http.MethodPost, "/flags", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	line := buf.String()
	if !strings.Contains(line, "500") {
		t.Fatalf("log line missing status 500: %q", line)
	}
}

func TestLoggingStripsUserQueryAndControlChars(t *testing.T) {
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	defer logger.SetOutput(os.Stdout)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := Logging(next)

	req := httptest.NewRequest(http.MethodGet, "/flags/evaluate?user=alice&x=1", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	line := buf.String()
	if strings.Contains(line, "user") || strings.Contains(line, "alice") {
		t.Fatalf("log line leaks user id: %q", line)
	}
	if !strings.Contains(line, "x=1") {
		t.Fatalf("log line should keep other query params: %q", line)
	}
}

func TestSanitizedPathStripsControlChars(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/flags/foo", nil)
	req.URL.Path = "/flags/foo\nbar\tbaz"
	req.URL.RawQuery = "note=a\rb"

	got := sanitizedPath(req)
	if strings.ContainsAny(got, "\n\r\t") {
		t.Fatalf("sanitized path still contains control chars: %q", got)
	}
}

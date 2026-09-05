package middleware

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var logger = log.New(os.Stdout, "", log.LstdFlags)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.ResponseWriter.Write(b)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		logger.Printf("%s %s %d %s", r.Method, sanitizedPath(r), rec.status, time.Since(start))
	})
}

func sanitizedPath(r *http.Request) string {
	path := r.URL.Path

	if r.URL.RawQuery != "" {
		q := r.URL.Query()
		q.Del("user")
		if len(q) > 0 {
			path = path + "?" + q.Encode()
		}
	}

	return stripControl(path)
}

func stripControl(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r < 0x20 || r == 0x7f {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

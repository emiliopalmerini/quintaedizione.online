package web

import (
	"compress/gzip"
	"net/http"
	"path/filepath"
	"strings"
)

// CompressionMiddleware gzip-compresses textual responses when requested.
func CompressionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Accept-Encoding")
		if r.Method == http.MethodHead || r.Header.Get("Range") != "" || !acceptsGzip(r) || !compressiblePath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		compressed := &gzipResponseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(compressed, r)
		compressed.finish()
	})
}

func acceptsGzip(r *http.Request) bool {
	for _, encoding := range strings.Split(r.Header.Get("Accept-Encoding"), ",") {
		if strings.TrimSpace(strings.SplitN(encoding, ";", 2)[0]) == "gzip" {
			return true
		}
	}
	return false
}

func compressiblePath(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case "", ".css", ".html", ".js", ".json", ".svg", ".txt", ".xml":
		return true
	default:
		return false
	}
}

type gzipResponseWriter struct {
	http.ResponseWriter
	writer  *gzip.Writer
	status  int
	started bool
}

func (w *gzipResponseWriter) WriteHeader(status int) {
	if w.started {
		return
	}
	w.status = status
	if status == http.StatusNoContent || status == http.StatusNotModified || status < 200 {
		w.started = true
		w.ResponseWriter.WriteHeader(status)
	}
}

func (w *gzipResponseWriter) Write(body []byte) (int, error) {
	w.start()
	if w.writer == nil {
		return w.ResponseWriter.Write(body)
	}
	return w.writer.Write(body)
}

func (w *gzipResponseWriter) Flush() {
	w.start()
	if w.writer != nil {
		_ = w.writer.Flush()
	}
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *gzipResponseWriter) start() {
	if w.started {
		return
	}
	w.started = true
	if w.Header().Get("Content-Encoding") == "" {
		w.Header().Del("Content-Length")
		w.Header().Set("Content-Encoding", "gzip")
		w.writer = gzip.NewWriter(w.ResponseWriter)
	}
	w.ResponseWriter.WriteHeader(w.status)
}

func (w *gzipResponseWriter) finish() {
	if !w.started {
		w.start()
	}
	if w.writer != nil {
		_ = w.writer.Close()
	}
}

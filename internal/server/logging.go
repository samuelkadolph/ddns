package server

import (
	"log"
	"net/http"
	"time"
)

func RequestLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := &LoggingResponseWriter{w: w}
		start := time.Now()

		handler.ServeHTTP(p, r)

		log.Printf("%s %s %s %s %v %d %d", r.Method, r.RequestURI, r.RemoteAddr, r.Header.Get("User-Agent"), time.Since(start), p.code, p.bytes)
	})
}

type LoggingResponseWriter struct {
	bytes int
	code  int
	w     http.ResponseWriter
}

func (w *LoggingResponseWriter) Header() http.Header {
	return w.w.Header()
}

func (w *LoggingResponseWriter) Write(b []byte) (int, error) {
	w.bytes += len(b)
	return w.w.Write(b)
}

func (w *LoggingResponseWriter) WriteHeader(c int) {
	w.code = c
	w.w.WriteHeader(c)
}

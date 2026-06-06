package lib

import (
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const defaultLogFile = "tmp/okane.log"

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func InitLogger() (*os.File, error) {
	logFile := os.Getenv("LOG_FILE")
	if logFile == "" {
		logFile = defaultLogFile
	}

	if err := os.MkdirAll(filepath.Dir(logFile), 0755); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	writer := io.MultiWriter(os.Stdout, file)
	logger := slog.New(slog.NewJSONHandler(writer, &slog.HandlerOptions{}))
	slog.SetDefault(logger)
	log.SetOutput(writer)

	return file, nil
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(wrapped, r)

		slog.Info(
			"request",
			"method", r.Method,
			"path", r.URL.Path,
			"query", r.URL.RawQuery,
			"status", wrapped.status,
			"duration", time.Since(start).String(),
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)
	})
}

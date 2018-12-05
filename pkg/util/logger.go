package util

import (
	"net/http"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// NewLogger creates a new logger
// TODO: pass log level
func NewLogger() (*zap.Logger, error) {
	l, err := zap.NewProduction()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to initialize zap logger")
	}
	return l, nil
}

// NewSugaredLogger creates a new sugared logger
func NewSugaredLogger() (*zap.SugaredLogger, error) {
	l, err := NewLogger()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to initialize zap logger")
	}
	return l.Sugar(), nil
}

// LoggerWrapper is a decorator for a HTTP Request, adding structured logging functionality
func LoggerWrapper(inner http.Handler, logger *zap.SugaredLogger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		inner.ServeHTTP(w, r)
		logger.Infow("Request completed",
			"remote-addr", r.RemoteAddr,
			"method", r.Method,
			"path", r.RequestURI,
			"duration", time.Since(start),
		)
	})
}

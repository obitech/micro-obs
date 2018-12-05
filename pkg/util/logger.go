package util

import (
	"net/http"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// NewLogger creates a new logger
// TODO: pass log level
func newLogger(level string) (*zap.Logger, error) {
	atom := zap.NewAtomicLevel()

	switch level {
	case "debug":
		atom.SetLevel(zap.DebugLevel)
	case "warn":
		atom.SetLevel(zap.WarnLevel)
	case "error":
		atom.SetLevel(zap.ErrorLevel)
	default:
		level = "info"
		atom.SetLevel(zap.InfoLevel)
	}

	cfg := zap.Config{
		Development:       false,
		DisableCaller:     true,
		DisableStacktrace: true,
		EncoderConfig:     zap.NewProductionEncoderConfig(),
		Encoding:          "json",
		ErrorOutputPaths:  []string{"stdout"},
		Level:             atom,
		OutputPaths:       []string{"stdout"},
	}
	l, err := cfg.Build()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to initialize zap Logger")
	}

	l.Debug("Logger created",
		zap.String("level", level),
	)
	return l, nil
}

// NewSugaredLogger creates a new sugared logger
func NewSugaredLogger(level string) (*zap.SugaredLogger, error) {
	l, err := newLogger(level)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to initialize zap SugaredLogger")
	}
	return l.Sugar(), nil
}

// LoggerWrapper is a decorator for a HTTP Request, adding structured logging functionality
func LoggerWrapper(inner http.Handler, logger *zap.SugaredLogger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		inner.ServeHTTP(w, r)
		logger.Infow("Request completed",
			"address", r.RemoteAddr,
			"method", r.Method,
			"path", r.RequestURI,
			"duration", time.Since(start),
		)
	})
}

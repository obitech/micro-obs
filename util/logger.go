package util

import (
	"net/http"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Logger is an adapter type for zap's SugaredLogger
type Logger struct {
	logger *zap.SugaredLogger
}

// Info uses fmt.Sprint to log a templated message.
func (l *Logger) Info(args ...interface{}) {
	l.Info(args...)
}

// Infof uses fmt.Sprintf to log a templated message.
func (l *Logger) Infof(msg string, args ...interface{}) {
	l.Infof(msg, args...)
}

// Infow logs a message with some additional context. The variadic key-value pairs are treated as they are in With.
func (l *Logger) Infow(msg string, kv ...interface{}) {
	l.Infow(msg, kv...)
}

func (l *Logger) Error(msg string) {
	l.Error(msg)
}

// newLogger creates a new logger
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
		DisableCaller:     false,
		DisableStacktrace: false,
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
		logger.Debugw("request received",
			"address", r.RemoteAddr,
			"method", r.Method,
			"path", r.RequestURI,
		)
		inner.ServeHTTP(w, r)
		logger.Infow("request completed",
			"address", r.RemoteAddr,
			"method", r.Method,
			"path", r.RequestURI,
			"duration", time.Since(start),
		)
	})
}

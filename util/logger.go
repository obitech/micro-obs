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
	l.logger.Info(args...)
}

// Infof uses fmt.Sprintf to log a templated message.
func (l *Logger) Infof(msg string, args ...interface{}) {
	l.logger.Infof(msg, args...)
}

// Infow logs a message with some additional context. The variadic key-value pairs are treated as they are in With.
func (l *Logger) Infow(msg string, kv ...interface{}) {
	l.logger.Infow(msg, kv...)
}

// Error uses fmt.Sprint to construct and log a message.
func (l *Logger) Error(msg string) {
	l.logger.Error(msg)
}

// Errorw logs a message with some additional context. The variadic key-value
// pairs are treated as they are in With.
func (l *Logger) Errorw(msg string, kv ...interface{}) {
	l.logger.Errorw(msg, kv...)
}

// Warn uses fmt.Sprint to log a templated message.
func (l *Logger) Warn(args ...interface{}) {
	l.logger.Warn(args...)
}

// Warnf uses fmt.Sprintf to log a templated message.
func (l *Logger) Warnf(msg string, args ...interface{}) {
	l.logger.Warnf(msg, args...)
}

// Warnw logs a message with some additional context. The variadic key-value pairs are treated as they are in With.
func (l *Logger) Warnw(msg string, kv ...interface{}) {
	l.logger.Warnw(msg, kv...)
}

// Debug uses fmt.Sprint to log a templated message.
func (l *Logger) Debug(args ...interface{}) {
	l.logger.Debug(args...)
}

// Debugf uses fmt.Sprintf to log a templated message.
func (l *Logger) Debugf(msg string, args ...interface{}) {
	l.logger.Debugf(msg, args...)
}

// Debugw logs a message with some additional context. The variadic key-value pairs are treated as they are in With.
func (l *Logger) Debugw(msg string, kv ...interface{}) {
	l.logger.Debugw(msg, kv...)
}

// Panic uses fmt.Sprint to log a templated message, then panics.
func (l *Logger) Panic(args ...interface{}) {
	l.logger.Panic(args...)
}

// Panicf uses fmt.Sprintf to log a templated message, then panics.
func (l *Logger) Panicf(msg string, args ...interface{}) {
	l.logger.Panicf(msg, args...)
}

// Panicw logs a message with some additional context, then panics The variadic key-value pairs are treated as they are in With.
func (l *Logger) Panicw(msg string, kv ...interface{}) {
	l.logger.Panicw(msg, kv...)
}

// Fatal uses fmt.Sprint to log a templated message, then calls os.Exit.
func (l *Logger) Fatal(args ...interface{}) {
	l.logger.Fatal(args...)
}

// Fatalf uses fmt.Sprintf to log a templated message, then calls os.Exit.
func (l *Logger) Fatalf(msg string, args ...interface{}) {
	l.logger.Fatalf(msg, args...)
}

// Fatalw logs a message with some additional context, then calls os.Exit. The variadic key-value pairs are treated as they are in With.
func (l *Logger) Fatalw(msg string, kv ...interface{}) {
	l.logger.Fatalw(msg, kv...)
}

// Sync flushes any buffered log entries.
func (l *Logger) Sync() {
	l.logger.Sync()
}

// NewLogger creates a new Logger.
func NewLogger(level string) (*Logger, error) {
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

	return &Logger{logger: l.Sugar()}, nil
}

// LoggerWrapper is a decorator for a HTTP Request, adding structured logging functionality
func LoggerWrapper(inner http.Handler, logger *Logger) http.HandlerFunc {
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

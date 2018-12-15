package item

import (
	"context"
	"github.com/obitech/micro-obs/util"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Server is a wrapper for a HTTP server, with dependencies attached.
type Server struct {
	address  string
	endpoint string
	redis    *redis.Client
	redisOps uint64
	server   *http.Server
	router   *mux.Router
	logger   *zap.SugaredLogger
}

// ServerOptions sets options when creating a new server.
type ServerOptions func(*Server) error

// NewServer creates a new Server according to options.
func NewServer(options ...ServerOptions) (*Server, error) {
	// Create default logger
	logger, err := util.NewSugaredLogger("info")
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create SugaredLogger")
	}

	// Sane defaults
	rc, _ := NewRedisClient("redis://127.0.0.1:6379/0")
	s := &Server{
		address:  ":8080",
		endpoint: "127.0.0.1:9090",
		redis:    rc,
		logger:   logger,
		router:   util.NewRouter(),
	}

	// Applying custom settings
	for _, fn := range options {
		if err := fn(s); err != nil {
			return nil, errors.Wrap(err, "Failed to set server options")
		}
	}

	// Instrumenting redis
	s.redis.WrapProcess(func(old func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
		return func(cmd redis.Cmder) error {
			atomic.AddUint64(&s.redisOps, 1)
			ops := atomic.LoadUint64(&s.redisOps)
			s.logger.Debugw("redis sent",
				"count", ops,
				"cmd", cmd,
			)
			err := old(cmd)
			s.logger.Debugw("redis received",
				"count", ops,
				"cmd", cmd,
			)
			return err
		}
	})

	s.logger.Debugw("Creating new server",
		"address", s.address,
		"endpoint", s.endpoint,
	)

	// Setting routes
	s.createRoutes()

	return s, nil
}

// NewRedisClient creates a new go-redis/redis client according to passed options.
// Address needs to be a valid redis URL, e.g. redis://127.0.0.1:6379/0 or redis://:qwerty@localhost:6379/1
func NewRedisClient(addr string) (*redis.Client, error) {
	opt, err := redis.ParseURL(addr)
	if err != nil {
		return nil, err
	}

	c := redis.NewClient(&redis.Options{
		Addr:     opt.Addr,
		Password: opt.Password,
		DB:       opt.DB,
	})

	return c, nil
}

// SetServerAddress sets the server address.
func SetServerAddress(address string) ServerOptions {
	return func(s *Server) error {
		if err := util.CheckTCPAddress(address); err != nil {
			return err
		}

		s.address = address
		return nil
	}
}

// SetServerEndpoint sets the server endpoint address for other services to call it.
func SetServerEndpoint(address string) ServerOptions {
	return func(s *Server) error {
		s.endpoint = address
		return nil
	}
}

// SetLogLevel sets the log level to either debug, warn, error or info. Info is default.
func SetLogLevel(level string) ServerOptions {
	return func(s *Server) error {
		l, err := util.NewSugaredLogger(level)
		if err != nil {
			return err
		}
		s.logger = l
		return nil
	}
}

// SetRedisAddress sets a custom address for the redis connection.
func SetRedisAddress(address string) ServerOptions {
	return func(s *Server) error {
		// Close old client
		err := s.redis.Close()
		if err != nil {
			s.logger.Warnw("Error while closing old redis client",
				"error", err,
			)
		}

		rc, err := NewRedisClient(address)
		if err != nil {
			return err
		}
		s.redis = rc
		return nil
	}
}

// Run starts a Server and shuts it down properly on a SIGINT and SIGTERM.
func (s *Server) Run() error {
	defer s.logger.Sync()
	defer s.redis.Close()

	// Checking for redis connection
	s.logger.Debug("Testing redis connection")
	_, err := s.redis.Ping().Result()
	if err != nil {
		return errors.Wrap(err, "Unable to connect to redis server")
	}

	// Create TCP listener
	l, err := net.Listen("tcp", s.address)
	if err != nil {
		return errors.Wrapf(err, "Failed creating listener on %s", s.address)
	}

	// Create HTTP Server
	s.server = &http.Server{
		Handler:        s.router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Adjusting custom settings
	go func() {
		s.logger.Infow("Server listening",
			"address", s.address,
			"endpoint", s.endpoint,
		)
		s.logger.Fatal(s.server.Serve(l))
	}()

	// Buffered channel to receive a single os.Signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Blocking channel until interrupt occurs
	<-stop
	s.Stop()

	return nil
}

// ServeHTTP dispatches the request to the matching mux handler.
// This function is mainly intended for testing purposes.
func (s *Server) serveHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) internalError(w http.ResponseWriter) {
	w.Header().Del("Content-Type")
	w.WriteHeader(http.StatusInternalServerError)
	if _, err := io.WriteString(w, "Internal Server Error\n"); err != nil {
		s.logger.Panicw("unable to send response",
			"error", err,
		)
	}
}

// Respond sends a JSON-encoded response.
func (s *Server) Respond(status int, m string, c int, data interface{}, w http.ResponseWriter) {
	res, err := util.NewResponse(status, m, c, data)
	if err != nil {
		s.internalError(w)
		s.logger.Panicw("unable to create JSON response",
			"error", err,
		)
	}

	err = res.SendJSON(w)
	if err != nil {
		s.internalError(w)
		s.logger.Panicw("sending JSON response failed",
			"error", err,
			"response", res,
		)
	}
}

// Stop will stop the server
func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	s.logger.Info("Shutting down")
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Errorw("HTTP server shutdown",
			"error", err,
		)
	}
}

package item

import (
	"github.com/obitech/micro-obs/util"
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Server is a wrapper for a HTTP server, with several additional information attached suche as a
// listener and endpoint address, a router and a logger.
// Since all routes hang off the Server, handlers can easily access the database session.
type Server struct {
	address  string
	endpoint string
	server   *http.Server
	router   *mux.Router
	logger   *zap.SugaredLogger
}

// ServerOptions sets options when creating a new server
type ServerOptions func(*Server) error

// NewServer creates a new Server according to options
func NewServer(options ...ServerOptions) (*Server, error) {
	// Create logger
	logger, err := util.NewSugaredLogger("info")
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create SugaredLogger")
	}

	// Sane defaults
	s := &Server{
		address:  ":8080",
		endpoint: "127.0.0.1:9090",
		logger: logger,
		router: util.NewRouter(),
	}

	// Setting passed server options
	for _, fn := range options {
		if err := fn(s); err != nil {
			return nil, errors.Wrap(err, "Failed to set server options")
		}
	}

	s.logger.Debugw("Creating new server",
		"address", s.address,
		"endpoint", s.endpoint,
	)

	// Setting routes
	s.createRoutes()

	return s, nil
}

// SetServerAddress sets the server address
func SetServerAddress(address string) ServerOptions {
	return func(s *Server) error {
		if err := util.CheckTCPAddress(address); err != nil {
			return err
		}

		s.address = address
		return nil
	}
}

// SetServerEndpoint sets the server endpoint address for other services to call it
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

// Run starts a Server and shuts it down properly on a SIGINT and SIGTERM
func (s *Server) Run() error {
	defer s.logger.Sync()

	// Create TCP listener
	l, err := net.Listen("tcp", s.address)
	if err != nil {
		return errors.Wrapf(err, "Failed creating listener on %s", s.address)
	}

	// Create HTTP Server
	s.server = &http.Server{
		Handler: s.router,
	}

	// Setting up goroutine for serving
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

// ServeHTTP dispatches the request to the matching mux handler
// This function is mainly intended for testing purposes
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// Stop will stop the server
func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	s.logger.Info("Shutting down")
	s.server.Shutdown(ctx)
}

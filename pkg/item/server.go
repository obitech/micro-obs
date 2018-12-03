package item

import (
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// Server defines holds information about routing and a zap Logger
// Since all routes hang off the Server, handlers can easily access the database session.
type Server struct {
	Router *mux.Router
	Logger *zap.SugaredLogger
}

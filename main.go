package main

import (
	"net/http"

	"github.com/obitech/micro-obs/pkg/item"
	"github.com/obitech/micro-obs/pkg/util"
)

func main() {
	logger, _ := util.NewSugaredLogger()
	defer logger.Sync()

	s := item.Server{
		Router: util.NewRouter(),
		Logger: logger,
	}

	s.Routes()
	addr := "127.0.0.1:8080"

	logger.Infow("Server listening",
		"addr", addr,
	)
	logger.Fatal(http.ListenAndServe(addr, s.Router))
}

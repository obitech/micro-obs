package main

import (
	"fmt"
	"os"

	"github.com/obitech/micro-obs/pkg/item"
)

func main() {
	s, err := item.NewServer(
		item.SetServerAddress(":8080"),
		item.SetServerEndpoint("127.0.0.1:8080"),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	if err := s.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}
}

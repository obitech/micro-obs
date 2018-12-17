package main

import (
	"fmt"
	"os"

	"github.com/obitech/micro-obs/order"
	"github.com/spf13/cobra"
)

func runServer(cmd *cobra.Command, args []string) {
	s, err := order.NewServer(
		order.SetServerAddress(address),
		order.SetServerEndpoint(endpoint),
		order.SetLogLevel(logLevel),
		order.SetRedisAddress(redis),
		order.SetItemServiceAddress(item),
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

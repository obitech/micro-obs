package main

import (
	"fmt"
	"os"

	"github.com/obitech/micro-obs/item"
	"github.com/spf13/cobra"
)

func runServer(cmd *cobra.Command, args []string) {
	s, err := item.NewServer(
		item.SetServerAddress(address),
		item.SetServerEndpoint(endpoint),
		item.SetLogLevel(logLevel),
		item.SetRedisAddress(redis),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	s.InitPromReg()

	if err := s.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(3)
	}
}

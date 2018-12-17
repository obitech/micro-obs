package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Default values to be used to initialize the order service
var (
	address  = ":9090"
	endpoint = "127.0.0.1:9091"
	logLevel = "info"
	redis    = "redis://127.0.0.1:6379/0"
	item     = "127.0.0.1:8080"
	rootCmd  = &cobra.Command{
		Use:   "order",
		Short: "Simple HTTP order serivce",
		Run:   runServer,
	}
)

// Execute runs the cobra rootCommand.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	f := rootCmd.Flags()
	f.StringVarP(&address, "address", "a", address, "listening address")
	f.StringVarP(&endpoint, "endpoint", "e", endpoint, "endpoint for other services to reach order service")
	f.StringVarP(&logLevel, "log-level", "l", logLevel, "log level (debug, info, warn, error), empty or invalid values will fallback to default")
	f.StringVarP(&redis, "redis-address", "r", redis, "redis address to connect to")
	f.StringVarP(&redis, "item-address", "i", item, "item service address to query")
}

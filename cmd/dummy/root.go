package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "dummy",
		Short: "populate item or order service with dummy data",
		Long:  "will contact either the item or order service via HTTP to populate data in the underlying datastores or send dummy requests",
	}
	allowedTargets = map[string]bool{"item": true, "order": true, "all": true}
	itemAddr       = "http://localhost:8080"
	orderAddr      = "http://localhost:8090"
)

func init() {
	rootCmd.AddCommand(dataCmd)
	rootCmd.AddCommand(requestsCmd)

	rootCmd.PersistentFlags().StringVarP(&itemAddr, "item-addr", "i", itemAddr, "address of the item service")
	rootCmd.PersistentFlags().StringVarP(&orderAddr, "order-addr", "o", orderAddr, "address of the order service")
}

// Execute runs the cobra rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		errExit(err)
	}
}

// Prints error and exits
func errExit(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

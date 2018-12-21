package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	orderAddr = "localhost:9090"
	rootCmd   = &cobra.Command{
		Use:   "dummy",
		Short: "create populate item or order service with dummy data",
		Long:  "will contact either the item or order service via HTTP to populate data in the underlying datastore",
	}
)

func init() {
	rootCmd.AddCommand(itemCmd)
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

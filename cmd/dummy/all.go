package main

import (
	"github.com/spf13/cobra"
)

var (
	allCmd = &cobra.Command{
		Use:   "all",
		Short: "target both the item and the order service",
	}

	dataAllCmd = &cobra.Command{
		Use:   "data",
		Short: "populate default data for the item service, followed by default data for the order service",
		Run:   populateAll,
	}

	requestsAllCmd = &cobra.Command{
		Use:   "requests",
		Short: "send 50 requests to each service",
		Run:   requestsAll,
	}
)

func init() {
	allCmd.AddCommand(dataAllCmd)
	allCmd.AddCommand(requestsAllCmd)
}

func populateAll(cmd *cobra.Command, args []string) {
	itemDefaultData(cmd, args)
	orderDefaultData(cmd, args)
}

func requestsAll(cmd *cobra.Command, args []string) {
	ch := make(chan bool, 2)

	go func() {
		itemRequests(cmd, args)
		ch <- true
	}()

	go func() {
		orderRequests(cmd, args)
		ch <- true
	}()

	<-ch
}

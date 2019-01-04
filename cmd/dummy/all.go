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
)

func init() {
	allCmd.AddCommand(dataAllCmd)
}

func populateAll(cmd *cobra.Command, args []string) {
	itemDefaultData(cmd, args)
	orderDefaultData(cmd, args)
}

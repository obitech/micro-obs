package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "dummy",
		Short: "populate item or order service with dummy data",
		Long:  "will contact either the item or order service via HTTP to populate data in the underlying datastores",
	}

	itemAddr  = "http://localhost:8080"
	orderAddr = "http://localhost:8090"
	numReq    = 15
	concReq   = 5
)

func init() {
	rootCmd.AddCommand(itemCmd)
	rootCmd.AddCommand(orderCmd)
	rootCmd.AddCommand(allCmd)

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

func sendRequests(addr string) {
	var ch chan bool
	limit := make(chan bool, concReq)

	if numReq == 0 {
		for {
			ch = make(chan bool)
			limit <- true
			go req(addr, ch, limit)
		}
	}
	ch = make(chan bool, numReq)
	for i := 0; i < numReq; i++ {
		limit <- true
		go req(addr, ch, limit)
	}
	<-ch
}

func req(addr string, ch chan bool, limit chan bool) {
	_, err := http.Get(addr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	time.Sleep(500 * time.Millisecond)
	<-limit
	ch <- true
}

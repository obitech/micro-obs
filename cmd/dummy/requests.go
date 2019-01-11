package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	requestsCmd = &cobra.Command{
		Use:   "requests [target] [handler]",
		Short: "sends requests to a handler of a single or all services",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return errors.New("requires two arguments: [target] [handler]")
			}

			if _, prs := allowedTargets[args[0]]; !prs {
				return errors.New(fmt.Sprintf("invalid argument, must be of [\"item\", \"order\", \"all\"]"))
			}

			return nil
		},
		Run: sendRequests,
	}
	numReq  = 15
	concReq = 5
)

func init() {
	requestsCmd.Flags().IntVarP(&numReq, "number", "n", 15, "number of requests to sent (0 for unlimited)")
	requestsCmd.Flags().IntVarP(&concReq, "concurrency", "c", 1, "number of requests to be sent concurrently")
}

func sendRequests(cmd *cobra.Command, args []string) {
	var url string

	switch args[0] {
	case "item":
		url = fmt.Sprintf("%s%s", itemAddr, args[1])
		concReqs(url)
	case "order":
		url = fmt.Sprintf("%s%s", orderAddr, args[1])
		concReqs(url)
	default:
		ch := make(chan bool, 2)

		url = fmt.Sprintf("%s%s", itemAddr, args[1])
		go func() {
			concReqs(url)
			ch <- true
		}()

		url = fmt.Sprintf("%s%s", orderAddr, args[1])
		go func() {
			concReqs(url)
			ch <- true
		}()

		<-ch
	}
}

func concReqs(url string) {
	var ch chan bool
	limit := make(chan bool, concReq)

	if numReq == 0 {
		for {
			ch = make(chan bool)
			limit <- true
			go req(url, ch, limit)
		}
	}

	ch = make(chan bool, numReq)
	for i := 0; i < numReq; i++ {
		limit <- true
		go req(url, ch, limit)
	}
	<-ch
}

func req(url string, ch chan bool, limit chan bool) {
	_, err := http.Get(url)
	errExit(err)

	time.Sleep(500 * time.Millisecond)
	<-limit
	ch <- true
}

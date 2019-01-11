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
		Use:   "requests [target] [handlers]",
		Short: "sends requests to handlers of a single or all services",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return errors.New("requires two arguments: [target] [handlers]")
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
	wait    = 500
)

func init() {
	requestsCmd.Flags().IntVarP(&numReq, "number", "n", numReq, "number of requests to sent (0 for unlimited)")
	requestsCmd.Flags().IntVarP(&concReq, "concurrency", "c", concReq, "number of requests to be sent concurrently")
	requestsCmd.Flags().IntVarP(&wait, "wait", "w", wait, "time to wait between requests in ms")
}

// TODO: worker pattern
func sendRequests(cmd *cobra.Command, args []string) {
	var url string
	var ch chan bool
	switch args[0] {
	case "item":
		for _, handler := range args[1:] {
			url = fmt.Sprintf("%s%s", itemAddr, handler)
			concReqs(url)
		}
	case "order":
		for _, handler := range args[1:] {
			url = fmt.Sprintf("%s%s", orderAddr, handler)
			concReqs(url)
		}
	default:
		ch = make(chan bool, 2*len(args[1:]))
		for _, handler := range args[1:] {
			url = fmt.Sprintf("%s%s", itemAddr, handler)
			go func() {
				concReqs(url)
				ch <- true
			}()
			url = fmt.Sprintf("%s%s", orderAddr, handler)
			go func() {
				concReqs(url)
				ch <- true
			}()
		}
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

	time.Sleep(time.Duration(wait) * time.Millisecond)
	<-limit
	ch <- true
}

package main

import (
	"fmt"
	"math/rand"
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
		Run: sendRequest,
	}
	numReq  = 15
	concReq = 2
	wait    = 500
)

func init() {
	requestsCmd.Flags().IntVarP(&numReq, "number", "n", numReq, "number of requests to sent")
	requestsCmd.Flags().IntVarP(&concReq, "concurrency", "c", concReq, "number of requests to be sent concurrently")
	requestsCmd.Flags().IntVarP(&wait, "wait", "w", wait, "time to wait between requests in ms")
}

func buildURLs(addr string, handlers ...string) []string {
	urls := []string{}

	// Total requests = -n * (number of handlers)
	for _, handler := range handlers {
		for j := 0; j < numReq; j++ {
			urls = append(urls, fmt.Sprintf("%s%s", addr, handler))
		}
	}

	// Shuffle slice so requests will be evenly distributed
	rand.Shuffle(len(urls), func(i, j int) {
		urls[i], urls[j] = urls[j], urls[i]
	})

	return urls
}

func sendRequest(cmd *cobra.Command, args []string) {
	switch args[0] {
	case "item":
		urls := buildURLs(itemAddr, args[1:]...)
		done := distributeWork(urls...)
		for i := 0; i < len(urls); i++ {
			<-done
		}

	case "order":
		urls := buildURLs(orderAddr, args[1:]...)
		done := distributeWork(urls...)
		for i := 0; i < len(urls); i++ {
			<-done
		}

	case "all":
		urls := []string{}
		urls = append(urls, buildURLs(itemAddr, args[1:]...)...)
		urls = append(urls, buildURLs(orderAddr, args[1:]...)...)
		rand.Shuffle(len(urls), func(i, j int) {
			urls[i], urls[j] = urls[j], urls[i]
		})

		done := distributeWork(urls...)
		for j := 0; j < len(urls); j++ {
			<-done
		}
	}
}

func distributeWork(urls ...string) <-chan bool {
	numJobs := len(urls)
	jobQ := make(chan string, numJobs)
	jobsDone := make(chan bool, numJobs)

	// Spawn -c workers
	for w := 0; w < concReq; w++ {
		go workerRequest(w, jobQ, jobsDone)
	}

	// Send requests to jobQ channel
	for _, url := range urls {
		jobQ <- url
	}

	// Work assigned, close jobQ channel
	close(jobQ)

	return jobsDone
}

func workerRequest(id int, jobs <-chan string, jobsDone chan<- bool) {
	// Grab jobs from channel and perform request
	for url := range jobs {
		start := time.Now()
		vl(fmt.Sprintf("Worker %d -> %s\n", id, url))

		_, err := http.Get(url)
		errExit(err)
		time.Sleep(time.Duration(wait) * time.Millisecond)

		vl(fmt.Sprintf("Worker %d <- Work to %s completed after %v\n", id, url, time.Since(start)))
		jobsDone <- true
	}
}

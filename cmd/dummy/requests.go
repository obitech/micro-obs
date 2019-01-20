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
		Short: "sends requests to handlers of a single or all services (targets)",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return errors.New("requires two arguments: [target] [handlers]")
			}

			if _, prs := allowedTargets[args[0]]; !prs {
				// Format allowed targets
				targets := make([]string, len(allowedTargets))
				i := 0
				for k := range allowedTargets {
					targets[i] = k
					i++
				}
				return errors.New(fmt.Sprintf("invalid argument, must be of %v", targets))
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

// buildURLS returns a shuffled slice of strings where the length is equal to the total number of requests.
// Each string is an URL and the slice represents the total amount of work to be handled by workers.
func buildURLs(addr string, handlers ...string) []string {
	urls := []string{}

	// Each handler is a single request to be made and gets added to the output slice
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

// sendRequest creates and distributes work according to passed arguments.
func sendRequest(cmd *cobra.Command, args []string) {
	switch args[0] {
	case "item":
		urls := buildURLs(itemAddr, args[1:]...)
		done := distributeRequests(urls...)
		for i := 0; i < len(urls); i++ {
			<-done
		}

	case "order":
		urls := buildURLs(orderAddr, args[1:]...)
		done := distributeRequests(urls...)
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

		done := distributeRequests(urls...)
		for j := 0; j < len(urls); j++ {
			<-done
		}
	}
}

// distributeRequests takes URLs as work, spwans workers according to flags and dsitributes work to them.
func distributeRequests(urls ...string) <-chan bool {
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

// workerRequest listenes on a passed channel for work, processes it, waits according to flags and then
// confirms job completion on a different chanel.
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

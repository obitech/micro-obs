package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	dataCmd = &cobra.Command{
		Use:   "data [target]",
		Short: "populate data to a single or all services",
		Long:  fmt.Sprintf("Sample data for item service: %s\nSample data for order service: %s", itemJSON, orderJSON),
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("requires one argument: [target]")
			}

			if _, prs := allowedTargets[args[0]]; !prs {
				return errors.New(fmt.Sprintf("invalid argument, must be of [\"item\", \"order\", \"all\"]"))
			}

			return nil
		},
		Run: sendData,
	}
	numDataReq  = 1
	concDataReq = 1
	waitData    = 0
)

func init() {
	dataCmd.Flags().IntVarP(&numDataReq, "number", "n", numDataReq, "number of times data should be sent")
	dataCmd.Flags().IntVarP(&concDataReq, "concurrency", "c", concDataReq, "number of requests to be sent concurrently")
	dataCmd.Flags().IntVarP(&waitData, "wait", "w", waitData, "time to wait between requests in ms")
}

type dataRequest struct {
	method string
	url    string
	data   []string
}

func sendData(cmd *cobra.Command, args []string) {
	switch args[0] {
	case "item":
		dr := dataRequest{
			method: "PUT",
			url:    fmt.Sprintf("%s/items", itemAddr),
			data:   itemJSON,
		}
		distributeWorkData(dr)

	case "order":
		dr := dataRequest{
			method: "POST",
			url:    fmt.Sprintf("%s/orders/create", orderAddr),
			data:   orderJSON,
		}
		distributeWorkData(dr)

	default:
		dr := dataRequest{
			method: "PUT",
			url:    fmt.Sprintf("%s/items", itemAddr),
			data:   itemJSON,
		}
		distributeWorkData(dr)

		dr = dataRequest{
			method: "POST",
			url:    fmt.Sprintf("%s/orders/create", orderAddr),
			data:   orderJSON,
		}
		distributeWorkData(dr)
	}
}

func distributeWorkData(dr dataRequest) {
	jobQ := make(chan dataRequest, numDataReq)
	jobsDone := make(chan bool, numDataReq)

	// Spwan -c workers
	for w := 0; w < concDataReq; w++ {
		go workerData(w, jobQ, jobsDone)
	}

	// Send -n data to jobQ channel
	for j := 0; j < numDataReq; j++ {
		jobQ <- dr
	}

	close(jobQ)

	// Retrieve
	for j := 0; j < numDataReq; j++ {
		<-jobsDone
	}
}

func workerData(id int, jobs <-chan dataRequest, jobsDone chan<- bool) {
	for dr := range jobs {
		start := time.Now()
		vl(fmt.Sprintf("Worker %d -> %s\n", id, dr.url))
		for _, js := range dr.data {
			buf := bytes.NewBuffer([]byte(js))
			req, err := http.NewRequest(dr.method, dr.url, buf)
			if err != nil {
				errExit(err)
			}
			req.Header.Add("Content-Type", "application/JSON; charset=UTF-8")

			vl(fmt.Sprintf("%s %s\n%s\n", dr.method, dr.url, dr.data))

			c := &http.Client{}
			res, err := c.Do(req)
			errExit(err)

			b, err := ioutil.ReadAll(res.Body)
			errExit(err)
			defer res.Body.Close()

			switch res.StatusCode {
			case http.StatusOK:
				fallthrough
			case http.StatusCreated:
				vl(fmt.Sprintf("%s", string(b)))
			default:
				fmt.Printf("Unexpected status code %d: %s", res.StatusCode, b)
				os.Exit(1)
			}
		}
		time.Sleep(time.Duration(waitData) * time.Millisecond)
		vl(fmt.Sprintf("Workder %d <- Work to %s completed after %v\n", id, dr.url, time.Since(start)))
		jobsDone <- true
	}
}

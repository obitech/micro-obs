package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var (
	orderCmd = &cobra.Command{
		Use:   "order",
		Short: "populate data for the order service",
		Run:   orderDefaultData,
	}

	orderPingCmd = &cobra.Command{
		Use:   "ping",
		Short: "ping the order service",
		Run:   orderPing,
	}

	orderDefaultCmd = &cobra.Command{
		Use:   "default",
		Short: "uses default data to populate",
		Long:  fmt.Sprintf("This command depends on the specific item IDs to be present in the item service. The following data will be sent to the service:\n%s", orderJSON),
		Run:   orderDefaultData,
	}

	orderRequestsCmd = &cobra.Command{
		Use:   "requests",
		Short: "sends requests to delay endpoint",
		Run:   orderRequests,
	}
)

func init() {
	orderRequestsCmd.Flags().IntVarP(&numReq, "count", "c", 15, "number of requests to send. 0 for unlimited.")

	orderCmd.AddCommand(orderPingCmd)
	orderCmd.AddCommand(orderDefaultCmd)
	orderCmd.AddCommand(orderRequestsCmd)
}

func orderPing(cmd *cobra.Command, args []string) {
	fmt.Printf("Pinging order service at %s\n", orderAddr)

	res, err := http.Get(orderAddr)
	errExit(err)

	b, err := ioutil.ReadAll(res.Body)
	errExit(err)
	defer res.Body.Close()

	fmt.Printf("order online: %s", string(b))
}

func orderDefaultData(cmd *cobra.Command, args []string) {
	method := "POST"
	url := fmt.Sprintf("%s/orders/create", orderAddr)
	data := orderJSON

	for _, js := range data {
		buf := bytes.NewBuffer([]byte(js))

		req, err := http.NewRequest(method, url, buf)
		if err != nil {
			errExit(err)
		}
		req.Header.Add("Content-Type", "application/JSON; charset=UTF-8")

		fmt.Printf("%s %s\n%s\n", method, url, js)

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
			fmt.Printf("%s", string(b))
		default:
			fmt.Printf("Unexpected status code %d: %s", res.StatusCode, b)
			os.Exit(1)
		}
	}
}

func orderRequests(cmd *cobra.Command, args []string) {
	addr := fmt.Sprintf("%s/delay", orderAddr)
	if numReq == 0 {
		for {
			sendRequest(addr)
		}
	}
	for i := 0; i < numReq; i++ {
		sendRequest(addr)
	}
}

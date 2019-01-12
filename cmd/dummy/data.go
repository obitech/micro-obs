package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var dataCmd = &cobra.Command{
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
	Run: populateData,
}

// TODO: number of times data should be populated, concurrently
func populateData(cmd *cobra.Command, args []string) {
	var url string
	var method string
	var data []string

	switch args[0] {
	case "item":
		method = "PUT"
		url = fmt.Sprintf("%s/items", itemAddr)
		data = itemJSON
		defaultData(method, url, data)
	case "order":
		method = "POST"
		url = fmt.Sprintf("%s/orders/create", orderAddr)
		data = orderJSON
		defaultData(method, url, data)
	default:
		method = "PUT"
		url = fmt.Sprintf("%s/items", itemAddr)
		data = itemJSON
		defaultData(method, url, data)

		method = "POST"
		url = fmt.Sprintf("%s/orders/create", orderAddr)
		data = orderJSON
		defaultData(method, url, data)
	}

}

func defaultData(method, url string, data []string) {
	for _, js := range data {
		buf := bytes.NewBuffer([]byte(js))

		req, err := http.NewRequest(method, url, buf)
		if err != nil {
			errExit(err)
		}
		req.Header.Add("Content-Type", "application/JSON; charset=UTF-8")

		vl(fmt.Sprintf("%s %s\n%s\n", method, url, js))

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

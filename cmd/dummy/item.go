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
	itemCmd = &cobra.Command{
		Use:   "item",
		Short: "populate data for the item service",
	}

	itemPingCmd = &cobra.Command{
		Use:   "ping",
		Short: "ping the item service",
		Run:   itemPing,
	}

	itemDefaultCmd = &cobra.Command{
		Use:   "default",
		Short: "uses default data to populate",
		Long:  fmt.Sprintf("the following data will be sent to the service:\n%s", itemJSON),
		Run:   itemDefaultData,
	}
)

func init() {
	itemCmd.AddCommand(itemPingCmd)
	itemCmd.AddCommand(itemDefaultCmd)
}

func itemPing(cmd *cobra.Command, args []string) {
	fmt.Printf("Pinging item service at %s\n", itemAddr)

	res, err := http.Get(itemAddr)
	errExit(err)

	b, err := ioutil.ReadAll(res.Body)
	errExit(err)
	defer res.Body.Close()

	fmt.Printf("item online: %s", string(b))
}

func itemDefaultData(cmd *cobra.Command, args []string) {
	method := "PUT"
	url := fmt.Sprintf("%s/items", itemAddr)
	data := itemJSON
	buf := bytes.NewBuffer([]byte(data))

	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		errExit(err)
	}
	req.Header.Add("Content-Type", "application/JSON; charset=UTF-8")

	fmt.Printf("%s %s\n%s\n", method, url, data)

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

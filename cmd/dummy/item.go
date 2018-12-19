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

	pingCmd = &cobra.Command{
		Use:   "ping",
		Short: "ping the item service",
		Run:   ping,
	}

	defaultCmd = &cobra.Command{
		Use:   "default",
		Short: "uses default data to populate",
		Long:  fmt.Sprintf("the following data will be sent to the item service:\n%s", itemJSON),
		Run:   defaultData,
	}

	itemAddr = "http://localhost:8080"
)

func init() {
	itemCmd.AddCommand(pingCmd)
	itemCmd.AddCommand(defaultCmd)

	pingCmd.Flags().StringVarP(&itemAddr, "addr", "a", itemAddr, "address of the of item service")
}

func ping(cmd *cobra.Command, args []string) {
	fmt.Printf("Pinging item service at %s\n", itemAddr)

	res, err := http.Get(itemAddr)
	errExit(err)

	b, err := ioutil.ReadAll(res.Body)
	errExit(err)
	defer res.Body.Close()

	fmt.Printf("item online: %s", string(b))
}

func defaultData(cmd *cobra.Command, args []string) {
	buf := bytes.NewBuffer([]byte(itemJSON))
	res, err := http.Post(fmt.Sprintf("%s/items", itemAddr), "application/JSON; charset=UTF-8", buf)
	errExit(err)

	b, err := ioutil.ReadAll(res.Body)
	errExit(err)
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
		fallthrough
	case http.StatusAccepted:
		fmt.Printf("%s", string(b))
	default:
		fmt.Printf("Unexpected status code %d: %s", res.StatusCode, b)
		os.Exit(1)
	}
}

package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

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

	itemAddr = "http://localhost:8080/"
)

func init() {
	itemCmd.AddCommand(pingCmd)

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

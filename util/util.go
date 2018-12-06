package util

import (
	"fmt"
	"net"
	"strconv"

	"github.com/pkg/errors"
)

// CheckTCPAddress checks if a passed TCP address is a valid listening address
func CheckTCPAddress(address string) error {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		fmt.Printf("Invalid host: %s\n", address)
	}

	ip := net.ParseIP(host)
	if ip == nil && host != "" {
		return errors.New(fmt.Sprintf("Error: Invalid IP address: %#v", host))
	}

	if err := CheckPort(port); err != nil {
		return err
	}
	return nil
}

// CheckPort checks if a port is in the valid 0 <= port <= 65535 range
func CheckPort(port string) error {
	var validServices = []string{"http"}
	for _, v := range validServices {
		if port == v {
			return nil
		}
	}

	i, err := strconv.Atoi(port)
	if err != nil {
		return errors.Wrapf(err, "Unable to convert %s to int", port)
	}

	if i < 0 || i > 65535 {
		return errors.New(fmt.Sprintf("Invalid port %s", port))
	}

	return nil
}

package util

import (
	"net"

	"github.com/pkg/errors"
)

// CheckAddress checks if a passed address is valid
func CheckAddress(address string) error {
	if _, _, err := net.SplitHostPort(address); err != nil {
		return errors.Wrapf(err, "Invalid address %s", address)
	}
	return nil
}

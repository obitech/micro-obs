package util

import (
	"fmt"
	"net"
	"strconv"

	"github.com/pkg/errors"
	"github.com/speps/go-hashids"
)

// CheckTCPAddress checks if a passed TCP address is a valid listening address
func CheckTCPAddress(address string) error {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		fmt.Printf("Invalid host: %s\n", address)
	}

	ip := net.ParseIP(host)
	if ip == nil && host != "" {
		return fmt.Errorf("Error: Invalid IP address: %#v", host)
	}

	if err := CheckPort(port); err != nil {
		return err
	}
	return err
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
		return fmt.Errorf("Invalid port %s", port)
	}

	return nil
}

// newHashID returns a HashID to perform en-/decoding on.
// See https://hashids.org for more info.
func newHashID() (*hashids.HashID, error) {
	// Defaults
	salt := "Best salt"
	minLength := 8

	// Initiliazing HashID
	hd := hashids.NewData()
	hd.Salt = salt
	hd.MinLength = minLength
	h, err := hashids.NewWithData(hd)
	return h, err
}

// StringToHashID converts a UTF-8 string into a HashID.
func StringToHashID(str string) (string, error) {
	h, err := newHashID()
	if err != nil {
		return "", err
	}

	// Convert string to []int
	var ss []int
	for _, v := range str {
		ss = append(ss, int(v))
	}

	// Encode
	e, err := h.Encode(ss)
	if err != nil {
		return "", err
	}
	return e, nil
}

// HashIDToString decodes a HashID-encoded string into a normal string
func HashIDToString(hash string) (string, error) {
	h, err := newHashID()
	if err != nil {
		return "", err
	}

	// Decode into []int
	ss, err := h.DecodeWithError(hash)
	if err != nil {
		return "", err
	}

	// Translate []int into string
	var s []rune
	for _, v := range ss {
		s = append(s, rune(v))
	}
	return string(s), nil
}

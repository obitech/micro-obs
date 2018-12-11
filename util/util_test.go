package util

import (
	"testing"
)

var (
	validAddr = []string{
		"127.0.0.1:8080",
		"0.0.0.0:80",
		":80",
		":8080",
		"127.0.0.1:80",
		"192.0.2.1:http",
	}

	invalidAddr = []string{
		":9999999",
		":-1",
		"asokdklasd",
		"0.0.0.0.0.0:80",
		"256.0.0.1:80",
	}

	validPort = []string{
		"0",
		"80",
		"42",
		"8080",
		"http",
	}

	invalidPort = []string{
		"9999999",
		"tcp",
		"-1",
	}

	toHash = []string{
		"Hello world",
		"0",
		"-1",
		"01234",
		"orange",
		"test",
		"😍",
		"👾 🙇 💁 🙅 🙆 🙋 🙎 🙍",
		"﷽",
		"     ",
		"1)4_?&$",
	}
)

func TestCheckTCPAddress(t *testing.T) {
	t.Run("Valid addresses", func(t *testing.T) {
		for _, tt := range validAddr {
			if err := CheckTCPAddress(tt); err != nil {
				t.Errorf("CheckTCPAddress(%#v) unsuccessful, expected: %v, got: %#v", tt, nil, err)
			}
		}
	})

	t.Run("Invalid addresses", func(t *testing.T) {
		for _, tt := range invalidAddr {
			if err := CheckTCPAddress(tt); err == nil {
				t.Errorf("CheckTCPAddress(%#v) should throw error, got: %v", tt, nil)
			}
		}
	})
}

func TestCheckPort(t *testing.T) {
	t.Run("Valid ports", func(t *testing.T) {
		for _, tt := range validPort {
			if err := CheckPort(tt); err != nil {
				t.Errorf("CheckPort(%#v) unsuccessful, expected: %v, got: %#v", tt, nil, err)
			}
		}
	})
}

func TestHashIDConversion(t *testing.T) {
	// String -> Encoded
	var hashMap = make(map[string]string)

	t.Run("Encoding", func(t *testing.T) {
		for _, tt := range toHash {
			h, err := StringToHashID(tt)
			if err != nil {
				t.Errorf("Unable to encode %#v to HashID: %#v", tt, err)
			}
			hashMap[tt] = h
		}
	})

	t.Run("Decoding", func(t *testing.T) {
		for k, v := range hashMap {
			s, err := HashIDToString(v)
			if err != nil {
				t.Errorf("Unable to decode %#v to string: %#v", v, err)
			}

			if s != k {
				t.Errorf("HashIDToString(%#v) unsuccsessful, expected: %#v, got: %#v", v, k, s)
			}
		}
	})
}

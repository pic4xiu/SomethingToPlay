package main

import (
	"github.com/malfunkt/iprange"
	"testing"
)

func FuzzReverse(f *testing.F) {
	testcases := []string{"10.0.0.1", "10.0.0.5-10", "192.168.1.*", "192.168.10.0/24"}
	for _, tc := range testcases {
		f.Add(tc) // Use f.Add to provide a seed corpus
	}
	f.Fuzz(func(t *testing.T, orig string) {
		_, err1 := iprange.ParseList(orig)
		if err1 != nil {
			return
		}
	})
}

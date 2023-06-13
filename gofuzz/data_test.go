package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"testing"
)

func FuzzReverse(f *testing.F) {
	// testcases := []string{"10.0.0.1/70", "10.0.0.5-10", "192.168.1.*", "192.168.10.0/24"}
	// for _, tc := range testcases {
	// 	f.Add(tc) // Use f.Add to provide a seed corpus
	// }
	f.Fuzz(func(t *testing.T, data []byte) {
		fmt.Println(data)
		file := bytes.NewReader(data)
		_, _, err1 := image.Decode(file)
		if err1 != nil {
			return
		}
	})
}

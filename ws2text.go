/*
Package main of Wordstar-to-text converter
*/
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	flag.Parse()

	args := flag.Args()

	// todo: check at least one arg
	// todo: check only one arg

	filename := args[0]

	fh, err := os.Open(filename)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer fh.Close()

	done := false

	for !done {
		bytes := make([]byte, 256)

		// read some bytes
		_, err := fh.Read(bytes)
		if err != nil {
			fmt.Println(err.Error())
			done = true
		}

		// print all bytes, strip high bit
		for _, b := range bytes {
			fmt.Printf("%c", b&0x7F)
		}
	}
}

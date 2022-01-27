/*
Package main of Wordstar-to-text converter
*/
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

func printable(b byte) bool {
	result := true

	// control characters are not printable
	if b < 32 {
		result = false
	}

	// but a TAB is printable
	if b == 9 {
		result = true
	}

	// and CR and LF are printable
	if b == 10 || b == 13 {
		result = true
	}

	return result
}

func main() {
	flag.Parse()

	args := flag.Args()

	// todo: check at least one arg
	// todo: check only one arg

	filename := args[0]

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	eof := false

	// print almost all bytes
	for _, b := range bytes {
		if b == 26 {
			// CTRL-Z indicates end of text file
			eof = true
		}

		if !eof {
			// strip high bit
			b2 := b & 0x07F

			// print printable characters (and TAB, CR, and LF)
			if printable(b2) {
				fmt.Printf("%c", b2)
			}
		}
	}
}

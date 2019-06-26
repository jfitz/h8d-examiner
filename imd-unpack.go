/*
 Package of main IMD unpacker
*/
package main

import (
	"flag"
	"fmt"
	"github.com/jfitz/h8d-examiner/utils"
	"os"
)

func main() {
	// parse command line options
	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		fmt.Println("No file specified")
		os.Exit(1)
	}

	// get file name
	fileName := args[0]

	// open the file
	fh, err := os.Open(fileName)
	utils.CheckAndExit(err)

	defer fh.Close()

	// read IMD header
	header := ""
	b := make([]byte, 1)

	for b[0] != 0x1a {
		_, err = fh.Read(b)
		utils.CheckAndExit(err)

		b0 := b[0]

		header += string(b0)
	}

	// display header
	fmt.Println(header)

	eof := false
	index := 0

	for !eof {
		// if index mod 10 == 0, skip 15 bytes
		if index%10 == 0 {
			// read extra data at the start of each 10-sector block
			postamble := make([]byte, 15)
			_, err = fh.Read(postamble)
			utils.CheckAndExit(err)
		}

		// read byte code
		_, err = fh.Read(b)
		utils.CheckAndExit(err)

		b0 := b[0]

		// validate byte code
		if b0 == 0x01 {
			// read 256 bytes and dump
			length := 256
			sector := make([]byte, length)
			_, err = fh.Read(sector)
			utils.CheckAndExit(err)

			utils.Dump(sector, index, "hex")
			index += 1
		} else if b0 == 0x02 {
			// read 1 byte and replicate 256 times and dump
			_, err = fh.Read(b)
			utils.CheckAndExit(err)

			fmt.Printf("Sector: %04XH (%d): (compressed)\n", index, index)
			index += 1
		} else {
			pos, err := fh.Seek(0, os.SEEK_CUR)
			utils.CheckAndExit(err)

			fmt.Printf("Unknown byte code %02X at position %04X\n", b0, pos)
			eof = true
		}
	}
}

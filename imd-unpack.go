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

	if len(args) < 2 {
		fmt.Println("Usage: imd-unpack source-file destination-file")
		os.Exit(1)
	}

	// get file names
	source_fileName := args[0]
	dest_filename := args[1]

	// open the files
	sfh, err := os.Open(source_fileName)
	utils.CheckAndExit(err)

	defer sfh.Close()

	dfh, err := os.Create(dest_filename)
	utils.CheckAndExit(err)

	// read IMD header
	header := ""
	b := make([]byte, 1)

	for b[0] != 0x1a {
		_, err = sfh.Read(b)
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
			_, err = sfh.Read(postamble)
			utils.CheckAndExit(err)
		}

		// read byte code
		_, err = sfh.Read(b)
		utils.CheckAndExit(err)

		b0 := b[0]

		// validate byte code
		if b0 == 0x01 {
			// read 256 bytes and dump
			length := 256
			sector := make([]byte, length)
			_, err = sfh.Read(sector)
			utils.CheckAndExit(err)

			dfh.Write(sector)
			index += 1
		} else if b0 == 0x02 {
			// read 1 byte and replicate 256 times and dump
			_, err = sfh.Read(b)
			utils.CheckAndExit(err)

			b0 = b[0]
			length := 256
			sector := make([]byte, length)
			for i := range sector {
				sector[i] = b0
			}

			dfh.Write(sector)
			index += 1
		} else {
			pos, err := sfh.Seek(0, os.SEEK_CUR)
			utils.CheckAndExit(err)

			fmt.Printf("Unknown byte code %02X at position %04X\n", b0, pos)
			eof = true
		}
	}
}

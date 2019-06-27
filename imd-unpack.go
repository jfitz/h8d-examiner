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

func read_normal_sector(fh *os.File) ([]byte, error) {
	// read 256 bytes and dump
	length := 256
	sector := make([]byte, length)

	_, err := fh.Read(sector)
	if err != nil {
		return []byte{}, err
	}

	return sector, nil
}

func read_compressed_sector(fh *os.File) ([]byte, error) {
	// read 1 byte and replicate 256 times and dump
	b := make([]byte, 1)
	_, err := fh.Read(b)
	if err != nil {
		return []byte{}, err
	}

	b0 := b[0]
	length := 256
	sector := make([]byte, length)
	for i := range sector {
		sector[i] = b0
	}

	return sector, nil
}

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

		// TODO: strip non-printable characters

		header += string(b0)
	}

	// display header
	fmt.Println(header)

	eof := false
	index := 0

	// TODO: detect EOF on sfh

	for !eof {
		// if index mod 10 == 0, read track header
		// TODO: use number of sectors to find next header (start with zero)
		if index%10 == 0 {
			// read track header
			header := make([]byte, 15)
			_, err = sfh.Read(header)
			utils.CheckAndExit(err)

			// side (0x01 or 0x02)
			// track (0x00 to 0x27)
			// 00
			// number of sectors (0x0A)
			// first sector (0x01)
			// sector map[number of sectors]
		}

		// TODO: read entire track, unmap sectors using map, then re-map for Heath H-17 CP/M sequence

		// read byte code
		_, err = sfh.Read(b)
		utils.CheckAndExit(err)

		b0 := b[0]

		// validate byte code
		if b0 == 0x01 {
			sector, err := read_normal_sector(sfh)
			utils.CheckAndExit(err)

			dfh.Write(sector)
			index += 1
		} else if b0 == 0x02 {
			sector, err := read_compressed_sector(sfh)
			utils.CheckAndExit(err)

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

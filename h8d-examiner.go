/*
Package main of virtual CPU runner
*/
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

func checkAndExit(e error) {
	if e != nil {
		fmt.Println(e.Error())
		os.Exit(1)
	}
}

func dumpSector(fh *os.File, sectorIndex int) error {
	sector := make([]byte, 256)

	pos := int64(sectorIndex) * 256

	_, err := fh.Seek(pos, 0)
	if err != nil {
		fmt.Println("Sector does not exist")
		return nil
	}

	_, err = fh.Read(sector)
	if err != nil {
		return err
	}

	fmt.Println("sector dump")

	if len(sector) != 256 {
		return errors.New("Invalid sector length")
	}

	for i := 0; i < len(sector); i += 16 {
		fmt.Printf("%02X: ", i)

		for j := 0; j < 16; j++ {
			index := i + j
			b := sector[index]
			fmt.Printf("%02X ", b)
		}

		for j := 0; j < 16; j++ {
			index := i + j
			b := sector[index]
			if b >= ' ' && b <= 127 {
				fmt.Printf("%c", b)
			} else {
				fmt.Print(".")
			}
		}
		fmt.Println()
	}

	return nil
}

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) == 0 {
		fmt.Println("No file specified")
		os.Exit(1)
	}

	fileName := args[0]

	// open file
	fh, err := os.Open(fileName)
	checkAndExit(err)

	defer fh.Close()

	sectorIndex := 0

	err = dumpSector(fh, sectorIndex)
	checkAndExit(err)

	sectorIndex = 1

	err = dumpSector(fh, sectorIndex)
	checkAndExit(err)
}

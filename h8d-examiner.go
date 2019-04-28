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

func readSectors(name string) ([]byte, error) {
	sector := make([]byte, 256)

	f, err := os.Open(name)
	if err != nil {
		return sector, err
	}

	defer f.Close()

	_, err = f.Read(sector)
	if err != nil {
		return sector, err
	}

	return sector, nil
}

func dumpSector(sector []byte) error {
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

func examine(moduleFile string) error {
	// read data
	sector, err := readSectors(moduleFile)
	if err != nil {
		return err
	}

	err = dumpSector(sector)
	if err != nil {
		return err
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

	moduleFile := args[0]

	err := examine(moduleFile)
	checkAndExit(err)
}

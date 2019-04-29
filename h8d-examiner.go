/*
Package main of virtual CPU runner
*/
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
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

	fmt.Println()

	if len(sector) != 256 {
		return errors.New("Invalid sector length")
	}

	fmt.Println()
	fmt.Printf("Sector: %04XH (%d):\n", sectorIndex, sectorIndex)
	fmt.Println()

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

	reader := bufio.NewReader(os.Stdin)

	// open file
	fh, err := os.Open(fileName)
	checkAndExit(err)

	defer fh.Close()

	sectorIndex := 0

	err = dumpSector(fh, sectorIndex)
	checkAndExit(err)

	for {
		fmt.Println()
		fmt.Printf(">")
		line, err := reader.ReadString('\n')
		checkAndExit(err)

		line = strings.TrimSpace(line)

		if line == "quit" {
			os.Exit(0)
		} else if line == "help" {
			fmt.Println("quit  - exit the program")
			fmt.Println("help  - print this message")
			fmt.Println("nnn   - dump sector nnn")
			fmt.Println("stats - display statistics")
		} else if line == "stats" {
			fmt.Printf("File: %s\n", fileName)
			fmt.Printf("Sector: %04XH (%d)\n", sectorIndex, sectorIndex)
		} else if line == "" {
			sectorIndex += 1

			err = dumpSector(fh, sectorIndex)
			checkAndExit(err)
		} else {
			fmt.Println("quit  - exit the program")
			fmt.Println("help  - print this message")
			fmt.Println("nnn   - dump sector nnn")
			fmt.Println("stats - display statistics")
		}
	}
}

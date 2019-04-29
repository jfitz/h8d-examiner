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
	"regexp"
	"strconv"
	"strings"
)

func checkAndExit(e error) {
	if e != nil {
		fmt.Println(e.Error())
		os.Exit(1)
	}
}

func dumpHex(bytes []byte) {
	for _, b := range bytes {
		fmt.Printf("%02X ", b)
	}
}

func dumpOctal(bytes []byte) {
	for _, b := range bytes {
		fmt.Printf("%03o ", b)
	}
}

func dumpAscii(bytes []byte) {
	for _, b := range bytes {
		if b >= ' ' && b <= 127 {
			fmt.Printf("%c", b)
		} else {
			fmt.Print(".")
		}
	}
}

func dumpSector(fh *os.File, sectorIndex int, base string) error {
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
		bytes := sector[i : i+16]

		fmt.Printf("%02X: ", i)

		if base == "hex" {
			dumpHex(bytes)
		} else {
			dumpOctal(bytes)
		}

		dumpAscii(bytes)

		fmt.Println()
	}

	return nil
}

func displayHelp() {
	fmt.Println("quit  - exit the program")
	fmt.Println("help  - print this message")
	fmt.Println("nnn   - dump sector nnn")
	fmt.Println("stats - display statistics")
	fmt.Println("octal - show dump in octal")
	fmt.Println("hex   - show dump in hex")
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

	base := "hex"
	sectorIndex := 0

	err = dumpSector(fh, sectorIndex, base)
	checkAndExit(err)

	numberPattern, err := regexp.Compile("^\\d+$")
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
			displayHelp()
		} else if line == "stats" {
			fmt.Printf("File: %s\n", fileName)
			fmt.Printf("Sector: %04XH (%d)\n", sectorIndex, sectorIndex)
		} else if line == "" {
			sectorIndex += 1

			err = dumpSector(fh, sectorIndex, base)
			checkAndExit(err)
		} else if numberPattern.MatchString(line) {
			sectorIndex, _ = strconv.Atoi(line)

			err = dumpSector(fh, sectorIndex, base)
			checkAndExit(err)
		} else if line == "octal" {
			base = "octal"

			err = dumpSector(fh, sectorIndex, base)
			checkAndExit(err)
		} else if line == "hex" {
			base = "hex"

			err = dumpSector(fh, sectorIndex, base)
			checkAndExit(err)
		} else {
			displayHelp()
		}
	}
}

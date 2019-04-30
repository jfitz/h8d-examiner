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

func dumpOctal(bytes []byte) {
	for _, b := range bytes {
		fmt.Printf(" %03o", b)
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
	// position at the desired sector
	pos := int64(sectorIndex) * 256

	_, err := fh.Seek(pos, 0)
	if err != nil {
		fmt.Println("Sector does not exist")
		return nil
	}

	// read the sector
	sector := make([]byte, 256)
	_, err = fh.Read(sector)
	if err != nil {
		return err
	}

	if len(sector) != 256 {
		return errors.New("Invalid sector length")
	}

	// display the sector
	fmt.Println()

	// print header information
	if base == "hex" {
		fmt.Printf("Sector: %04XH (%d):\n", sectorIndex, sectorIndex)
	} else {
		highByte := sectorIndex / 256
		lowByte := sectorIndex % 256
		fmt.Printf("Sector: %03o.%03oA (%d):\n", highByte, lowByte, sectorIndex)
	}

	fmt.Println()

	// print data in lines of 16 bytes
	for i := 0; i < len(sector); i += 16 {
		bytes := sector[i : i+16]

		// print in hex or octal
		if base == "hex" {
			fmt.Printf("%02X: ", i)
		} else {
			fmt.Printf("%03o: ", i)
		}

		if base == "hex" {
			fmt.Printf("% 02X", bytes)
		} else {
			dumpOctal(bytes)
		}

		fmt.Print("  ")

		// print in ASCII (with dots for non-printable bytes)
		dumpAscii(bytes)

		fmt.Println()
	}

	return nil
}

func displayHelp() {
	fmt.Println("stats - display statistics")
	fmt.Println("hdos  - interpret as HDOS disk")
	fmt.Println("cp/m  - interpret as CP/M disk")
	fmt.Println("quit  - exit the program")
	fmt.Println("help  - print this message")
}

func displaySectorHelp() {
	fmt.Println("exit  - exit to main level")
	fmt.Println("<RET> - dump next sector")
	fmt.Println("nnn   - dump sector nnn")
	fmt.Println("octal - show dump in octal")
	fmt.Println("hex   - show dump in hex")
	fmt.Println("help  - print this message")
}

func sector(reader *bufio.Reader, fh *os.File) {
	// set default values
	base := "hex"
	sectorIndex := 0
	lastWasDump := false

	numberPattern, err := regexp.Compile("^\\d+$")
	checkAndExit(err)

	// display the first sector
	err = dumpSector(fh, sectorIndex, base)
	checkAndExit(err)
	fmt.Println()
	lastWasDump = true

	// prompt for command and process it
	for {
		// display prompt and read command
		fmt.Printf("SECTOR> ")
		line, err := reader.ReadString('\n')
		checkAndExit(err)

		// process the command
		line = strings.TrimSpace(line)

		if line == "exit" {
			return
		} else if line == "" {
			if lastWasDump {
				sectorIndex += 1
			}

			err = dumpSector(fh, sectorIndex, base)
			checkAndExit(err)
			fmt.Println()
			lastWasDump = true
		} else if numberPattern.MatchString(line) {
			sectorIndex, _ = strconv.Atoi(line)

			err = dumpSector(fh, sectorIndex, base)
			checkAndExit(err)
			fmt.Println()
			lastWasDump = true
		} else if line == "octal" {
			base = "octal"

			err = dumpSector(fh, sectorIndex, base)
			checkAndExit(err)
			fmt.Println()
			lastWasDump = true
		} else if line == "hex" {
			base = "hex"

			err = dumpSector(fh, sectorIndex, base)
			checkAndExit(err)
			fmt.Println()
			lastWasDump = true
		} else {
			displaySectorHelp()
			fmt.Println()
		}
	}
}

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

	reader := bufio.NewReader(os.Stdin)

	// open the file
	fh, err := os.Open(fileName)
	checkAndExit(err)

	defer fh.Close()

	// get file statistics
	fileInfo, err := fh.Stat()
	fileSize := fileInfo.Size()
	fileSizeInK := fileSize / 1024
	fileSectorCount := fileSize / 256
	fileLastSector := fileSectorCount - 1

	// prompt for command and process it
	for {
		// display prompt and read command
		fmt.Printf("> ")
		line, err := reader.ReadString('\n')
		checkAndExit(err)

		// process the command
		line = strings.TrimSpace(line)

		if line == "quit" {
			os.Exit(0)
		} else if line == "stats" {
			fmt.Printf("File: %s\n", fileName)
			fmt.Printf("Size: %d (%dK)\n", fileSize, fileSizeInK)
			fmt.Printf("Last sector: %04XH (%d)\n", fileLastSector, fileLastSector)
			fmt.Println()
		} else if line == "sector" {
			sector(reader, fh)
		} else if line == "hdos" {
			fmt.Println("not implemented")
		} else if line == "cp/m" {
			fmt.Println("not implemented")
		} else if line == "help" {
			displayHelp()
			fmt.Println()
		} else {
			displayHelp()
			fmt.Println()
		}
	}
}

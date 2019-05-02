/*
Package main of H-8/H-89 disk reader
*/
package main

import (
	"bufio"
	"bytes"
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

func readSector(fh *os.File, sectorIndex int) ([]byte, error) {
	sector := make([]byte, 256)

	// position at the desired sector
	pos := int64(sectorIndex) * 256

	_, err := fh.Seek(pos, 0)
	if err != nil {
		return sector, errors.New("Sector does not exist")
	}

	// read the sector
	_, err = fh.Read(sector)
	if err != nil {
		return sector, err
	}

	if len(sector) != 256 {
		return sector, errors.New("Invalid sector length")
	}

	return sector, nil
}

func dumpSector(fh *os.File, sectorIndex int, base string) error {
	sector, err := readSector(fh, sectorIndex)
	if err != nil {
		return err
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

func displayHdosHelp() {
	fmt.Println("stats - display statistics")
	fmt.Println("cat   - list files on disk")
	fmt.Println("dir   - same as CAT")
	fmt.Println("type  - display contents of file")
	fmt.Println("dump  - dump contents of file")
	fmt.Println("copy  - copy file to your filesystem")
	fmt.Println("exit  - exit to main level")
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

func hdos(reader *bufio.Reader, fh *os.File) {
	// read sector 9
	sectorIndex := 9
	sector, err := readSector(fh, sectorIndex)
	if err != nil {
		fmt.Println("Cannot read sector 9")
		return
	}

	// extract and validate sector number for first directory sector [10,399]
	dis := int(sector[3]) + int(sector[4])*256

	// extract and validate sector number for GRT [10,399]
	grt := int(sector[5]) + int(sector[6])*256

	// extract and validate sectors per group in 2,4,8
	spg := int(sector[7])

	// extract and validate init.abs version in 00h,15h,16h,20h
	ver := int(sector[9])

	siz := 400
	pss := 256
	if ver > 0x20 {
		// extract and validate number of sectors
		siz = int(sector[12]) + int(sector[13])*256

		// extract and validate sector size == 256
		pss = int(sector[14]) + int(sector[15])*256

		// extract and validate flags 0 => 40tk1s 1=> 40tk2s 2=> 80tk1s 3=> 80tk2s
	}

	// extract and validate label ASCII text, zero terminated
	labelBytes := sector[17:77]
	n := bytes.IndexByte(labelBytes, byte(0))
	if n > -1 {
		labelBytes = labelBytes[:n]
	}

	labelError := false
	for _, b := range labelBytes {
		if b < 32 || b > 126 {
			labelError = true
		}
	}

	label := string(labelBytes)

	// extract and validate sectors per track == 10
	spt := int(sector[79])

	// if version 20h: num sectors match flags 0 => 400 1 => 800 2 => 800 3 => 1600

	if labelError {
		fmt.Println("This disk has a strange label")
	}

	// prompt for command and process it
	done := false
	for !done {
		// display prompt and read command
		fmt.Printf("HDOS> ")
		line, err := reader.ReadString('\n')
		checkAndExit(err)

		// process the command
		line = strings.TrimSpace(line)

		if line == "exit" {
			done = true
		} else if line == "stats" {
			fmt.Printf("First directory sector: 0x%02X (%d)\n", dis, dis)
			fmt.Printf("GRT sector: 0x%02X (%d)\n", grt, grt)
			fmt.Printf("Sectors per group: %d\n", spg)
			fmt.Printf("INIT.ABS version: 0x%02X\n", ver)
			fmt.Printf("Number of sectors: %d\n", siz)
			fmt.Printf("Sector size: %d\n", pss)
			fmt.Printf("Sectors per track: %d\n", spt)
			fmt.Printf("Label: %s\n", label)
			fmt.Println()
		} else if line == "cat" || line == "dir" {
			fmt.Println("not implemented")
		} else if line == "type" {
			fmt.Println("not implemented")
		} else if line == "dump" {
			fmt.Println("not implemented")
		} else if line == "copy" {
			fmt.Println("not implemented")
		} else if line == "help" {
			displayHdosHelp()
			fmt.Println()
		} else {
			displayHdosHelp()
			fmt.Println()
		}
	}
}

func cpm(reader *bufio.Reader, fh *os.File) {
	fmt.Println("not implemented")
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
			hdos(reader, fh)
		} else if line == "cp/m" {
			cpm(reader, fh)
		} else if line == "help" {
			displayHelp()
			fmt.Println()
		} else {
			displayHelp()
			fmt.Println()
		}
	}
}

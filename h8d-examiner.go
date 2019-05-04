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

func trimSlice(slice []byte) []byte {
	n := bytes.IndexByte(slice, byte(0))

	if n > -1 {
		slice = slice[:n]
	}

	return slice
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

func readSectorPair(fh *os.File, sectorIndex int) ([]byte, error) {
	// read 2 sectors (512 bytes)
	first, err := readSector(fh, sectorIndex)
	if err != nil {
		return []byte{}, errors.New("Cannot read first directory sector")
	}

	second, err := readSector(fh, sectorIndex+1)
	if err != nil {
		return []byte{}, errors.New("Cannot read second directory sector")
	}

	directoryBlock := append(first, second...)

	return directoryBlock, nil
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

func dateToText(dateBytes []byte) string {
	monthNames := [12]string{"JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"}

	day := int((dateBytes[0] & 0xF8) >> 3)
	month := int((dateBytes[0]&0x03)<<1) + int((dateBytes[1]&0x80)>>7)
	monthName := monthNames[month]
	year := int((dateBytes[1]&0x7E)>>1) + 1970

	s := fmt.Sprintf("%02d-%s-%02d", day, monthName, year)

	return s
}

func flagsToText(flags byte) string {
	text := ""

	if (flags & 0200) == 0200 {
		text += "S"
	} else {
		text += " "
	}

	if (flags & 0100) == 0100 {
		text += "L"
	} else {
		text += " "
	}

	if (flags & 0040) == 0040 {
		text += "W"
	} else {
		text += " "
	}

	if (flags & 0020) == 0020 {
		text += "C"
	} else {
		text += " "
	}

	return text
}

func getSectors(grt []byte, firstCluster byte, lastCluster byte, lastSector int, sectorsPerGroup int) []int {
	sectors := []int{}

	index := firstCluster
	max := sectorsPerGroup

	need_one := true
	for index != 0 || need_one {
		if index == lastCluster {
			max = lastSector
		}

		for i := 0; i < max; i++ {
			sector := int(index)*sectorsPerGroup + i
			sectors = append(sectors, sector)
		}

		index = grt[index]
		need_one = false
	}

	return sectors
}

func printDirectoryBlock(directoryBlock []byte, grtSector []byte, sectorsPerGroup int) {
	// parse and print 22 entries of 23 bytes each
	for i := 0; i < 22; i++ {
		start := i * 23
		end := start + 23
		entry := directoryBlock[start:end]
		nameBytes := entry[0:8]
		if nameBytes[0] < 0xfe {
			extensionBytes := entry[8:11]
			name := string(trimSlice(nameBytes))
			extension := string(trimSlice(extensionBytes))
			flagByte := entry[14]
			flags := flagsToText(flagByte)

			createDateBytes := entry[19:21]
			createDate := dateToText(createDateBytes)
			modifyDateBytes := entry[21:23]
			modifyDate := dateToText(modifyDateBytes)

			firstCluster := entry[16]
			lastCluster := entry[17]
			lastSector := int(entry[18])
			usedSectors := getSectors(grtSector, firstCluster, lastCluster, lastSector, sectorsPerGroup)
			usedSectorCount := len(usedSectors)
			allocSectors := getSectors(grtSector, firstCluster, lastCluster, sectorsPerGroup, sectorsPerGroup)
			allocSectorCount := len(allocSectors)

			fmt.Printf("%-8s.%-3s    %s     %s    %s   %3d   %3d\n", name, extension, flags, createDate, modifyDate, usedSectorCount, allocSectorCount)
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

	// default values for H-17 SSSD disk
	siz := 400
	pss := 256
	spt := 10

	// HDOS 2.0 knows about H-47 and H-37 disks
	if ver > 0x20 {
		// extract and validate number of sectors
		siz = int(sector[12]) + int(sector[13])*256

		// extract and validate sector size == 256
		pss = int(sector[14]) + int(sector[15])*256

		// extract and validate flags 0 => 40tk1s 1=> 40tk2s 2=> 80tk1s 3=> 80tk2s

		// extract and validate sectors per track == 10
		spt = int(sector[79])
	}

	// extract and validate label ASCII text, zero terminated
	labelBytes := trimSlice(sector[17:77])

	labelError := false
	for _, b := range labelBytes {
		if b < 32 || b > 126 {
			labelError = true
		}
	}

	label := string(labelBytes)

	// if version 20h: num sectors match flags 0 => 400 1 => 800 2 => 800 3 => 1600

	if labelError {
		fmt.Println("This disk has a strange label")
	}

	grtSector, err := readSector(fh, grt)
	checkAndExit(err)

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

			freeSectors := getSectors(grtSector, 0, 0, spg, spg)
			freeSectorCount := len(freeSectors)
			fmt.Printf("Free sectors: %d\n", freeSectorCount)
			fmt.Println()
		} else if line == "cat" || line == "dir" {
			fmt.Println("Name            Flags    Created        Modified      Used  Allocated")

			// start with first directory sector
			sectorIndex := dis

			for sectorIndex != 0 {
				directoryBlock, err := readSectorPair(fh, sectorIndex)
				if err != nil {
					fmt.Println(err.Error())
				}

				printDirectoryBlock(directoryBlock, grtSector, spg)

				// read 6 bytes
				vectorBytes := directoryBlock[506:512]

				// bytes [4] and [5] are index of next directory pair
				sectorIndex = int(vectorBytes[4]) + int(vectorBytes[5])*256
			}
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
			fmt.Printf("Image: %s\n", fileName)
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

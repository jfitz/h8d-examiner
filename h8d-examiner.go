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

func mainHelp() {
	fmt.Println("stats - display statistics")
	fmt.Println("hdos  - interpret as HDOS disk")
	fmt.Println("cp/m  - interpret as CP/M disk")
	fmt.Println("quit  - exit the program")
}

func sectorHelp() {
	fmt.Println("exit  - exit to main level")
	fmt.Println("<RET> - dump next sector")
	fmt.Println("nnn   - dump sector nnn")
	fmt.Println("octal - show dump in octal")
	fmt.Println("hex   - show dump in hex")
}

func hdosHelp() {
	fmt.Println("stats - display statistics")
	fmt.Println("cat   - list files on disk")
	fmt.Println("dir   - same as CAT")
	fmt.Println("type  - display contents of file")
	fmt.Println("dump  - dump contents of file")
	fmt.Println("copy  - copy file to your filesystem")
	fmt.Println("exit  - exit to main level")
}

func cpmHelp() {
	fmt.Println("stats - display statistics")
	fmt.Println("dir   - list files on disk")
	fmt.Println("type  - display contents of file")
	fmt.Println("dump  - dump contents of file")
	fmt.Println("copy  - copy file to your filesystem")
	fmt.Println("exit  - exit to main level")
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

	fileInfo, err := fh.Stat()
	fileSize := fileInfo.Size()
	fileSizeInK := fileSize / 1024
	fileSectorCount := fileSize / 256
	fileLastSector := fileSectorCount - 1

	// prompt for command and process it
	done := false
	for !done {
		// display prompt and read command
		fmt.Printf("SECTOR> ")
		line, err := reader.ReadString('\n')
		checkAndExit(err)

		// process the command
		line = strings.TrimSpace(line)

		if line == "exit" {
			fmt.Println()
			done = true
		} else if line == "stats" {
			fmt.Printf("Size: %d (%dK)\n", fileSize, fileSizeInK)
			fmt.Printf("Last sector: %04XH (%d)\n", fileLastSector, fileLastSector)
			fmt.Println()
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
			sectorHelp()
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

			fmt.Printf("%-8s.%-3s    %s     %s    %s   %4d   %4d\n", name, extension, flags, createDate, modifyDate, usedSectorCount, allocSectorCount)
		}
	}
}

type HdosLabel struct {
	Dis  int
	Grt  int
	Spg  int
	Ver  int
	Siz  int
	Pss  int
	Spt  int
	Text string
}

func (label *HdosLabel) Init(sector []byte) {
	// extract and validate sector number for first directory sector [10,399]
	label.Dis = int(sector[3]) + int(sector[4])*256

	// extract and validate sector number for GRT [10,399]
	label.Grt = int(sector[5]) + int(sector[6])*256

	// extract and validate sectors per group in 2,4,8
	label.Spg = int(sector[7])

	// extract and validate init.abs version in 00h,15h,16h,20h
	label.Ver = int(sector[9])

	// default values for H-17 SSSD disk
	label.Siz = 400
	label.Pss = 256
	label.Spt = 10

	// HDOS 2.0 knows about H-47 and H-37 disks
	if label.Ver > 0x20 {
		// extract and validate number of sectors
		label.Siz = int(sector[12]) + int(sector[13])*256

		// extract and validate sector size == 256
		label.Pss = int(sector[14]) + int(sector[15])*256

		// extract and validate flags 0 => 40tk1s 1=> 40tk2s 2=> 80tk1s 3=> 80tk2s

		// extract and validate sectors per track == 10
		label.Spt = int(sector[79])
	}

	// extract and validate label ASCII text, zero terminated
	labelBytes := trimSlice(sector[17:77])

	label.Text = string(labelBytes)

	// if version 20h: num sectors match flags 0 => 400 1 => 800 2 => 800 3 => 1600
}

func (label HdosLabel) Print() {
	fmt.Printf("First directory sector: 0x%02X (%d)\n", label.Dis, label.Dis)
	fmt.Printf("GRT sector: 0x%02X (%d)\n", label.Grt, label.Grt)
	fmt.Printf("Sectors per group: %d\n", label.Spg)
	fmt.Printf("INIT.ABS version: 0x%02X\n", label.Ver)
	fmt.Printf("Number of sectors: %d\n", label.Siz)
	fmt.Printf("Sector size: %d\n", label.Pss)
	fmt.Printf("Sectors per track: %d\n", label.Spt)
	fmt.Printf("Label: %s\n", label.Text)
}

func hdosDir(fh *os.File, hdosLabel HdosLabel, grtSector []byte) {
	fmt.Println("Name            Flags    Created        Modified      Used  Allocated")

	// start with first directory sector
	sectorIndex := hdosLabel.Dis

	for sectorIndex != 0 {
		directoryBlock, err := readSectorPair(fh, sectorIndex)
		if err != nil {
			fmt.Println(err.Error())
		}

		printDirectoryBlock(directoryBlock, grtSector, hdosLabel.Spg)

		// read 6 bytes
		vectorBytes := directoryBlock[506:512]

		// bytes [4] and [5] are index of next directory pair
		sectorIndex = int(vectorBytes[4]) + int(vectorBytes[5])*256
	}

	fmt.Println()
}

func hdos(reader *bufio.Reader, fh *os.File) {
	// read sector 9
	sectorIndex := 9
	sector, err := readSector(fh, sectorIndex)
	if err != nil {
		fmt.Println("Cannot read sector 9")
		return
	}

	hdosLabel := HdosLabel{}
	hdosLabel.Init(sector)

	// check text label
	labelError := false
	for _, c := range hdosLabel.Text {
		if c < 32 || c > 126 {
			labelError = true
		}
	}

	if labelError {
		fmt.Println("This disk has a strange label")
	}

	// read Group Reservation Table (GRT)
	grtSector, err := readSector(fh, hdosLabel.Grt)
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
			fmt.Println()
			done = true
		} else if line == "stats" {
			hdosLabel.Print()

			freeSectors := getSectors(grtSector, 0, 0, hdosLabel.Spg, hdosLabel.Spg)
			freeSectorCount := len(freeSectors)
			fmt.Printf("Free sectors: %d\n", freeSectorCount)
			fmt.Println()
		} else if line == "cat" || line == "dir" {
			hdosDir(fh, hdosLabel, grtSector)
		} else if line == "type" {
			fmt.Println("not implemented")
		} else if line == "dump" {
			fmt.Println("not implemented")
		} else if line == "copy" {
			fmt.Println("not implemented")
		} else {
			hdosHelp()
			fmt.Println()
		}
	}
}

func cpmRecords(block int, dirBase int) []int {
	recordMap := [][]int{
		{0, 1, 8, 9, 16, 17, 4, 5},
		{12, 13, 2, 3, 10, 11, 18, 19},
		{6, 7, 14, 15, 20, 21, 28, 29},
		{36, 37, 24, 25, 32, 33, 22, 23},
		{30, 31, 38, 39, 26, 27, 34, 35},
	}

	index := block % 5
	return recordMap[index]
}

type SectorAndOffset struct {
	Sector int
	Offset int
}

func (sectorAndOffset SectorAndOffset) to_string() string {
	return fmt.Sprintf("%d:%d", sectorAndOffset.Sector, sectorAndOffset.Offset)
}

func cpmRecordToSectorAndOffset(record int) SectorAndOffset {
	sector := record / 2
	offset := record % 2
	sectorAndOffset := SectorAndOffset{sector, offset}

	return sectorAndOffset
}

func cpmDir(fh *os.File, directory []byte) {
	fmt.Println("Name          Extent Flags User Records")

	index := 0
	entrySize := 32

	for index < len(directory) {
		end := index + entrySize
		entry := directory[index:end]

		user := int(entry[0])
		nameBytes := entry[1:9]
		if nameBytes[0] >= 32 && nameBytes[0] <= 126 {
			extensionBytes := [3]byte{}
			extensionBytes[0] = entry[9] & 0x7F
			extensionBytes[1] = entry[10] & 0x7F
			extensionBytes[2] = entry[11] & 0x7F

			extent := int(entry[12])

			recordCount := int(entry[15])

			allocationBytes := trimSlice(entry[16:32])
			blocks := []int{}
			for _, b := range allocationBytes {
				blocks = append(blocks, int(b))
			}

			// extract flags from extension
			flag1Bit := (entry[9] & 0x80) == 0x80
			flag2Bit := (entry[10] & 0x80) == 0x80
			flag3Bit := (entry[11] & 0x80) == 0x80

			// convert bytes to strings
			name := string(trimSlice(nameBytes))
			extension := string(trimSlice(extensionBytes[:]))

			flags := ""
			if flag1Bit {
				flags += "W"
			} else {
				flags += " "
			}
			if flag2Bit {
				flags += "S"
			} else {
				flags += " "
			}
			if flag3Bit {
				flags += "A"
			} else {
				flags += " "
			}

			fmt.Printf("%-8s.%-3s    %2d    %s  %3d    %4d", name, extension, extent, flags, user, recordCount)

			records := []int{}
			for _, block := range blocks {
				blockRecords := cpmRecords(block, 60)
				records = append(records, blockRecords...)
			}

			records = records[:recordCount]

			fmt.Printf(" Blocks: % 02X\n", blocks)

			for _, record := range records {
				sectorAndOffset := cpmRecordToSectorAndOffset(record)
				fmt.Printf("%s ", sectorAndOffset.to_string())
			}
			fmt.Println()

			fmt.Println()
		}

		index += entrySize
	}

	fmt.Println()
}

func cpm(reader *bufio.Reader, fh *os.File) {
	// read sector 30 and 34
	sectorIndex := 30
	sector1, err := readSector(fh, sectorIndex)
	if err != nil {
		fmt.Println("Cannot read sector 30")
		return
	}

	sectorIndex = 34
	sector2, err := readSector(fh, sectorIndex)
	if err != nil {
		fmt.Println("Cannot read sector 34")
		return
	}

	directory := append(sector1, sector2...)

	// prompt for command and process it
	done := false
	for !done {
		// display prompt and read command
		fmt.Printf("CP/M> ")
		line, err := reader.ReadString('\n')
		checkAndExit(err)

		// process the command
		line = strings.TrimSpace(line)

		if line == "exit" {
			fmt.Println()
			done = true
		} else if line == "stats" {
			fmt.Println("not implemented")
		} else if line == "dir" {
			cpmDir(fh, directory)
		} else if line == "type" {
			fmt.Println("not implemented")
		} else if line == "dump" {
			fmt.Println("not implemented")
		} else if line == "copy" {
			fmt.Println("not implemented")
		} else {
			cpmHelp()
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
			fmt.Println()
			os.Exit(0)
		} else if line == "stats" {
			fmt.Printf("Image: %s\n", fileName)
			fmt.Printf("Size: %d (%dK)\n", fileSize, fileSizeInK)
			fmt.Printf("Last sector: %04XH (%d)\n", fileLastSector, fileLastSector)
			fmt.Println()
		} else if line == "sector" {
			fmt.Println()
			sector(reader, fh)
		} else if line == "hdos" {
			fmt.Println()
			hdos(reader, fh)
		} else if line == "cp/m" {
			fmt.Println()
			cpm(reader, fh)
		} else {
			mainHelp()
			fmt.Println()
		}
	}
}

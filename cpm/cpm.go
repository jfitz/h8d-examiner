/*
Package cpm of H-8/H-89 disk reader
*/
package cpm

import (
	"bufio"
	"fmt"
	"github.com/jfitz/h8d-examiner/utils"
	"os"
	"strings"
)

func help() {
	fmt.Println("stats - display statistics")
	fmt.Println("cat   - list directory entries and details")
	fmt.Println("dir   - list files on disk")
	fmt.Println("type  - display contents of file")
	fmt.Println("dump  - dump contents of file")
	fmt.Println("copy  - copy file to your filesystem")
	fmt.Println("exit  - exit to main level")
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

func stripHighBit(bs []byte) []byte {
	result := make([]byte, len(bs))

	for i, b := range bs {
		result[i] = b & 0x7F
	}

	return result
}

func getHighBit(bs []byte) []bool {
	result := make([]bool, len(bs))

	for i, b := range bs {
		result[i] = (b & 0x80) == 0x80
	}

	return result
}

// convert extension flags (the 'normal' ones) to text
func flagsToText(flags []bool) string {
	text := ""

	if flags[0] {
		text += "W"
	} else {
		text += " "
	}

	if flags[1] {
		text += "S"
	} else {
		text += " "
	}

	if flags[2] {
		text += "A"
	} else {
		text += " "
	}

	return text
}

// convert name flags to text
func specialFlagsToText(flags []bool) string {
	text := ""

	for i := 0; i < 8; i++ {
		if flags[i] {
			text += fmt.Sprintf("%d", i+1)
		} else {
			text += " "
		}
	}

	return text
}

func allRecords(blocks []int, directoryFirstRecord int, recordCount int) []int {
	records := []int{}

	for _, block := range blocks {
		blockRecords := cpmRecords(block, directoryFirstRecord)
		records = append(records, blockRecords...)
	}

	records = records[:recordCount]

	return records
}

func recordsToText(records []int) string {
	text := ""

	for _, record := range records {
		sectorAndOffset := cpmRecordToSectorAndOffset(record)
		text += fmt.Sprintf("%s ", sectorAndOffset.to_string())
	}

	return text
}

// print detailed catalog from directory
func cpmCat(fh *os.File, directory []byte) {
	fmt.Println("User Name          Extent Flags         Records")

	index := 0
	entrySize := 32
	directoryFirstRecord := 60

	for index < len(directory) {
		end := index + entrySize
		entry := directory[index:end]

		user := int(entry[0])

		// todo: user 0-31 else print alternate format
		// todo: entry outside 32-126 print alternate format
		if entry[1] >= 32 && entry[1] <= 126 {
			// normal directory entry

			nameBytes := stripHighBit(entry[1:9])
			name := string(utils.TrimSlice(nameBytes))

			extensionBytes := stripHighBit(entry[9:12])
			extension := string(utils.TrimSlice(extensionBytes))

			extent := int(entry[12])

			recordCount := int(entry[15])

			// extract flags from extension and name
			name_flags := getHighBit(entry[1:9])
			extension_flags := getHighBit(entry[9:12])

			// convert bytes to strings
			flags := flagsToText(extension_flags) + specialFlagsToText(name_flags)

			// print the information
			fmt.Printf("%3d  %-8s.%-3s    %2d    %s    %4d", user, name, extension, extent, flags, recordCount)

			// diag: print blocks
			allocationBytes := utils.TrimSlice(entry[16:32])
			blocks := []int{}
			for _, b := range allocationBytes {
				blocks = append(blocks, int(b))
			}

			fmt.Printf(" Blocks: % 02X\n", blocks)

			// diag: print record numbers
			records := allRecords(blocks, directoryFirstRecord, recordCount)

			text := recordsToText(records)
			fmt.Println(text)

			fmt.Println()
		} else {
			// strange directory entry - probably empty or deleted

			nameBytes := stripHighBit(entry[1:9])
			name := string(utils.TrimSlice(nameBytes))

			extensionBytes := stripHighBit(entry[9:12])
			extension := string(utils.TrimSlice(extensionBytes))

			// print the information
			fmt.Printf("%3d  %-8s.%-3s", user, name, extension)
		}

		index += entrySize
	}

	fmt.Println()
}

// print file-oriented directory (one line per file, not per entry)
func cpmDir(fh *os.File, directory []byte) {
	fmt.Println("Name          Extent Flags User Records")

	// for each user (0 to 31)
	for user := 0; user < 32; user++ {
		// get list of all file names with no repeats (strip flags)
		// if any
		// for each file
		// get all allocation blocks in order
		// convert allocation blocks to CP/M records
		// convert records to sector-offset pairs
		// print user, file name, and sector-offset pairs
	}

	fmt.Println()
}

func Menu(reader *bufio.Reader, fh *os.File) {
	// read sector 30 and 34
	sectorIndex := 30
	sector1, err := utils.ReadSector(fh, sectorIndex)
	if err != nil {
		fmt.Println("Cannot read sector 30")
		return
	}

	sectorIndex = 34
	sector2, err := utils.ReadSector(fh, sectorIndex)
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
		utils.CheckAndExit(err)

		// process the command
		line = strings.TrimSpace(line)

		if line == "exit" {
			fmt.Println()
			done = true
		} else if line == "stats" {
			fmt.Println("not implemented")
		} else if line == "cat" {
			cpmCat(fh, directory)
		} else if line == "dir" {
			cpmDir(fh, directory)
		} else if line == "type" {
			fmt.Println("not implemented")
		} else if line == "dump" {
			fmt.Println("not implemented")
		} else if line == "copy" {
			fmt.Println("not implemented")
		} else {
			help()
			fmt.Println()
		}
	}
}

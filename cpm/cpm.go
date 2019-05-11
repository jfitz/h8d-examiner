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

// print detailed catalog from directory
func cpmCat(fh *os.File, directory []byte) {
	fmt.Println("Name          Extent Flags         User Records")

	index := 0
	entrySize := 32

	for index < len(directory) {
		end := index + entrySize
		entry := directory[index:end]

		user := int(entry[0])
		nameBytes := entry[1:9]
		if nameBytes[0] >= 32 && nameBytes[0] <= 126 {
			extensionBytes := stripHighBit(entry[9:12])

			extent := int(entry[12])

			recordCount := int(entry[15])

			allocationBytes := utils.TrimSlice(entry[16:32])
			blocks := []int{}
			for _, b := range allocationBytes {
				blocks = append(blocks, int(b))
			}

			// extract flags from extension and name
			extension_flags := getHighBit(entry[9:12])
			name_flags := getHighBit(nameBytes)

			// convert bytes to strings
			name := string(utils.TrimSlice(stripHighBit(nameBytes)))
			extension := string(utils.TrimSlice(extensionBytes))

			// convert extension flags (the 'normal' ones) to text
			flags := ""
			if extension_flags[0] {
				flags += "W"
			} else {
				flags += " "
			}
			if extension_flags[1] {
				flags += "S"
			} else {
				flags += " "
			}
			if extension_flags[2] {
				flags += "A"
			} else {
				flags += " "
			}

			// convert name flags to text
			for i := 0; i < 8; i++ {
				if name_flags[i] {
					flags += fmt.Sprintf("%d", i+1)
				} else {
					flags += " "
				}
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

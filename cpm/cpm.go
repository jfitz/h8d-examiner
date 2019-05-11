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

			allocationBytes := utils.TrimSlice(entry[16:32])
			blocks := []int{}
			for _, b := range allocationBytes {
				blocks = append(blocks, int(b))
			}

			// extract flags from extension
			flag1Bit := (entry[9] & 0x80) == 0x80
			flag2Bit := (entry[10] & 0x80) == 0x80
			flag3Bit := (entry[11] & 0x80) == 0x80

			// convert bytes to strings
			name := string(utils.TrimSlice(nameBytes))
			extension := string(utils.TrimSlice(extensionBytes[:]))

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

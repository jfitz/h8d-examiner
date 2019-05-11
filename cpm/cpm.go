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
	fmt.Println("cat   - list directory entries")
	fmt.Println("cats  - list directory entries and details")
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

type DirectoryEntry struct {
	User        byte     // 0
	Name        [8]byte  // 1-8
	Extension   [3]byte  // 9-11
	Extent      byte     // 12
	S1          byte     // 13
	S2          byte     // 14
	RecordCount byte     // 15
	Blocks      [16]byte // 16-31
}

func (entry *DirectoryEntry) Init(bs []byte) {
	// user
	entry.User = bs[0]

	// name
	for i := 0; i < 8; i++ {
		entry.Name[i] = bs[1+i]
	}

	// extension
	for i := 0; i < 3; i++ {
		entry.Extension[i] = bs[9+i]
	}

	// extent
	entry.Extent = bs[12]

	// s1 and s2
	entry.S1 = bs[13]
	entry.S2 = bs[14]

	// record count
	entry.RecordCount = bs[15]

	// allocation blocks
	for i := 0; i < 16; i++ {
		entry.Blocks[i] = bs[16+i]
	}
}

func (entry DirectoryEntry) NormalName() bool {
	return entry.Name[1] >= 32 && entry.Name[1] <= 126
}

func (entry DirectoryEntry) ToText() string {
	if entry.NormalName() {
		return entry.normalToText()
	}

	return entry.deletedEntryToText()
}

// normal directory entry
func (entry DirectoryEntry) normalToText() string {
	user := int(entry.User)

	nameBytes := stripHighBit(entry.Name[:])
	name := string(utils.TrimSlice(nameBytes))

	extensionBytes := stripHighBit(entry.Extension[:])
	extension := string(utils.TrimSlice(extensionBytes))

	extent := int(entry.Extent)

	recordCount := int(entry.RecordCount)

	// extract flags from extension and name
	name_flags := getHighBit(entry.Name[:])
	extension_flags := getHighBit(entry.Extension[:])

	// convert bytes to strings
	flags := flagsToText(extension_flags) + specialFlagsToText(name_flags)

	// print the information
	text := fmt.Sprintf("%3d  %-8s.%-3s    %2d    %s    %4d", user, name, extension, extent, flags, recordCount)

	return text
}

// strange directory entry - probably empty or deleted
func (entry DirectoryEntry) deletedEntryToText() string {
	user := int(entry.User)

	nameBytes := stripHighBit(entry.Name[:])
	name := string(utils.TrimSlice(nameBytes))

	extensionBytes := stripHighBit(entry.Extension[:])
	extension := string(utils.TrimSlice(extensionBytes))

	// print the information
	text := fmt.Sprintf("%3d  %-8s.%-3s", user, name, extension)

	return text
}

func (entry DirectoryEntry) AllocationBlocks() []int {
	allocationBytes := utils.TrimSlice(entry.Blocks[:])

	blocks := []int{}

	for _, b := range allocationBytes {
		blocks = append(blocks, int(b))
	}

	return blocks
}

// print detailed catalog from directory
func cpmCat(fh *os.File, directory []byte, details bool) {
	fmt.Println("User Name          Extent Flags         Records")

	index := 0
	entrySize := 32
	directoryFirstRecord := 60

	for index < len(directory) {
		end := index + entrySize
		entry := directory[index:end]
		e2 := DirectoryEntry{}
		e2.Init(entry)

		// user := int(entry[0])

		// todo: user 0-31 else print alternate format
		// todo: entry outside 32-126 print alternate format
		text := e2.ToText()
		fmt.Println(text)

		if details {
			// diag: print block numbers and record numbers
			if e2.NormalName() {
				// block numbers
				blocks := e2.AllocationBlocks()
				fmt.Printf(" Blocks: % 02X\n", blocks)

				// record numbers
				recordCount := int(e2.RecordCount)
				records := allRecords(blocks, directoryFirstRecord, recordCount)

				recordText := recordsToText(records)
				fmt.Println(recordText)
			}
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
			cpmCat(fh, directory, false)
		} else if line == "cats" {
			cpmCat(fh, directory, true)
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

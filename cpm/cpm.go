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
	fmt.Println("stats  - display statistics")
	fmt.Println("cat    - list directory entries")
	fmt.Println("cats   - list directory entries and details")
	fmt.Println("dir    - list files on disk")
	fmt.Println("type   - display contents of file")
	fmt.Println("dump   - dump contents of file")
	fmt.Println("export - copy file to your filesystem")
	fmt.Println("exit   - exit to main level")
}

func blockToRecordsH17(block int, dirBase int) []int {
	recordMap := [][]int{
		{0, 1, 8, 9, 16, 17, 4, 5},
		{12, 13, 2, 3, 10, 11, 18, 19},
		{6, 7, 14, 15, 20, 21, 28, 29},
		{36, 37, 24, 25, 32, 33, 22, 23},
		{30, 31, 38, 39, 26, 27, 34, 35},
	}

	index := block % 5
	track := (block / 5) * 40
	offsets := recordMap[index]

	records := []int{}
	for _, offset := range offsets {
		record := dirBase + track + offset

		records = append(records, record)
	}

	return records
}

func blockToRecordsH37(block int, dirBase int) []int {
	recordMap := [][]int{
		{0, 1, 6, 7, 12, 13, 18, 19},
		{4, 5, 10, 11, 16, 17, 2, 3},
		{8, 9, 14, 15, 20, 21, 26, 27},
		{32, 33, 38, 39, 24, 25, 30, 31},
		{36, 37, 22, 23, 28, 29, 34, 35},
	}

	index := block % 5
	track := (block / 5) * 40
	offsets := recordMap[index]

	records := []int{}
	for _, offset := range offsets {
		record := dirBase + track + offset

		records = append(records, record)
	}

	return records
}

type SectorAndOffset struct {
	Sector int
	Offset int
}

func (sectorAndOffset SectorAndOffset) to_string() string {
	return fmt.Sprintf("%d:%d", sectorAndOffset.Sector, sectorAndOffset.Offset)
}

func recordToSectorAndOffset(record int) SectorAndOffset {
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

func allRecords(blocks []int, directoryFirstRecord int, recordCount int, diskParams utils.DiskParams) []int {
	records := []int{}

	if diskParams.Type == 1 {
		for _, block := range blocks {
			blockRecords := blockToRecordsH37(block, directoryFirstRecord)
			records = append(records, blockRecords...)
		}
	}

	if diskParams.Type == 0 {
		for _, block := range blocks {
			blockRecords := blockToRecordsH17(block, directoryFirstRecord)
			records = append(records, blockRecords...)
		}
	}

	records = records[:recordCount]

	return records
}

func recordsToText(records []int) string {
	text := ""

	for _, record := range records {
		sectorAndOffset := recordToSectorAndOffset(record)
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

func (entry DirectoryEntry) normalName() bool {
	byte1 := entry.Name[1] & 0x7F

	return byte1 >= 32 && byte1 <= 126
}

func (entry DirectoryEntry) normalExtent() bool {
	byte1 := entry.Extent

	return byte1 <= 0x80
}

func (entry DirectoryEntry) nameToText() string {
	name := strings.TrimSpace(string(stripHighBit(entry.Name[:])))
	extension := strings.TrimSpace(string(stripHighBit(entry.Extension[:])))
	filename := name + "." + extension

	return filename
}

func (entry DirectoryEntry) toText() string {
	if entry.normalName() && entry.normalExtent() {
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

func (entry DirectoryEntry) allocationBlocks() []int {
	allocationBytes := utils.TrimSlice(entry.Blocks[:])

	blocks := []int{}

	for _, b := range allocationBytes {
		blocks = append(blocks, int(b))
	}

	return blocks
}

// print detailed catalog from directory
func catCommand(fh *os.File, directory []byte, details bool, diskParams utils.DiskParams) {
	fmt.Println("User Name          Extent Flags         Records Blocks")

	index := 0
	entrySize := 32
	directoryFirstRecord := 60

	for index < len(directory) {
		end := index + entrySize
		entry := DirectoryEntry{}
		entry.Init(directory[index:end])

		// user := int(entry[0])

		// todo: user 0-31 print normal format
		// todo: user 0xE5 print deleted format
		// todo: else print alternate format
		text := entry.toText()
		fmt.Print(text)

		// print block numbers and maybe record numbers
		if entry.normalName() && entry.normalExtent() {
			// block numbers
			blocks := entry.allocationBlocks()
			fmt.Printf("   %02X", blocks)

			if details {
				// record numbers
				fmt.Println()
				recordCount := int(entry.RecordCount)
				records := allRecords(blocks, directoryFirstRecord, recordCount, diskParams)

				recordText := recordsToText(records)
				fmt.Println(recordText)
			}
		}

		fmt.Println()

		index += entrySize
	}

	fmt.Println()
}

// print file-oriented directory (one line per file, not per entry)
func dirCommand(fh *os.File, directory []byte, diskParams utils.DiskParams) {
	fmt.Println("Name          Flags      Records")

	// for each user (0 to 31)
	for user := 0; user < 32; user++ {
		// get list of all file names with no repeats (strip flags)
		index := 0
		entrySize := 32
		directoryFirstRecord := 60

		fileBlocks := map[string]int{}
		fileFlags := map[string]string{}

		for index < len(directory) {
			end := index + entrySize
			entry := DirectoryEntry{}
			entry.Init(directory[index:end])

			entryUser := int(entry.User)

			if entryUser == user {
				// get filename
				filename := entry.nameToText()

				if entry.Extent == 0 {
					// extract flags from extension and name
					name_flags := getHighBit(entry.Name[:])
					extension_flags := getHighBit(entry.Extension[:])
					flags := flagsToText(extension_flags) + specialFlagsToText(name_flags)
					fileFlags[filename] = flags
				}

				// calculate size
				blocks := entry.allocationBlocks()
				recordCount := int(entry.RecordCount)
				records := allRecords(blocks, directoryFirstRecord, recordCount, diskParams)
				fileBlocks[filename] += len(records)
			}

			index += entrySize
		}

		// for each file, print info
		for filename, size := range fileBlocks {
			flags := fileFlags[filename]
			fmt.Printf("%-12s  %s %5d\n", filename, flags, size)
		}
	}

	fmt.Println()
}

func readRecord(fh *os.File, recordNumber int) ([]byte, error) {
	sectorNumber := recordNumber / 2
	offset := recordNumber % 2

	recordBytes := []byte{}
	sectorBytes, err := utils.ReadSector(fh, sectorNumber)
	if err != nil {
		return recordBytes, err
	}

	start := 0 + 128*offset
	end := start + 128
	recordBytes = sectorBytes[start:end]

	return recordBytes, nil
}

func displayText(bytes []byte) {
	seenCtrlZ := false

	for _, b := range bytes {
		if b == 0x1A {
			seenCtrlZ = true
		} else {
			if !seenCtrlZ {
				fmt.Print(string(b))
			}
		}
	}
}

func displayRecords(fh *os.File, recordNumbers []int) {
	// for each record in block
	for _, record := range recordNumbers {
		// read data
		recordBytes, err := readRecord(fh, record)

		if err != nil {
			fmt.Println("Could not read record")
		} else {
			// print data
			displayText(recordBytes)
		}
	}
}

func dumpRecords(fh *os.File, recordNumbers []int) {
	// for each record in block
	for i, record := range recordNumbers {
		fmt.Printf("RECORD: %d\n", i)
		// read data
		recordBytes, err := readRecord(fh, record)

		if err != nil {
			fmt.Println("Could not read record")
		} else {
			// print data
			utils.Dump(recordBytes, i, "hex")
			fmt.Println()
		}
	}
}

func exportRecords(fh *os.File, recordNumbers []int, filename string, exportDirectory string) {
	fmt.Println("Exporting file...")

	// open file
	exportFilename := exportDirectory + "/" + filename
	f, err := os.Create(exportFilename)
	defer f.Close()

	if err != nil {
		fmt.Println("Cannot open file")
		return
	}

	// for each record in block
	for _, record := range recordNumbers {

		// read data
		recordBytes, err := readRecord(fh, record)

		if err != nil {
			fmt.Println("Could not read record")
		} else {
			// print data
			f.Write(recordBytes)
		}
	}

	fmt.Println("Done")
}

func getRecordNumbers(fh *os.File, directory []byte, user int, name string, extension string, diskParams utils.DiskParams) ([]int, bool) {
	recordNumbers := []int{}

	entrySize := 32
	directoryFirstRecord := 60

	recordsPerBlock := 128
	done := false

	anyFound := false
	extent := 0
	filename := name + "." + extension
	for extent < 128 && !done {
		index := 0
		found := false

		for index < len(directory) {
			end := index + entrySize
			entry := DirectoryEntry{}
			entry.Init(directory[index:end])

			if int(entry.User) == user && entry.nameToText() == filename && int(entry.Extent) == extent {
				found = true

				blocks := entry.allocationBlocks()
				recordCount := int(entry.RecordCount)

				// assume that the last block has a record count less than recordsPerBlock
				if recordCount < recordsPerBlock {
					done = true
				}

				blockRecordNumbers := allRecords(blocks, directoryFirstRecord, recordCount, diskParams)
				recordNumbers = append(recordNumbers, blockRecordNumbers...)
			}

			index += entrySize
		}

		if found {
			anyFound = true
			extent += 1
		} else {
			done = true
		}
	}

	return recordNumbers, anyFound
}

func splitFilename(filename string) (int, string, string) {
	// split filename into user, file, and name
	parts := strings.Split(filename, ".")
	name := parts[0]
	// todo: file may have no extension
	extension := parts[1]
	// todo: split user from filename
	user := 0

	return user, name, extension
}

func typeCommand(fh *os.File, directory []byte, filename string, diskParams utils.DiskParams) {
	user, name, extension := splitFilename(filename)

	recordNumbers, found := getRecordNumbers(fh, directory, user, name, extension, diskParams)

	if found {
		displayRecords(fh, recordNumbers)
	} else {
		fmt.Println("File not found")
	}

	fmt.Println()
	fmt.Println()
}

func dumpCommand(fh *os.File, directory []byte, filename string, diskParams utils.DiskParams) {
	user, name, extension := splitFilename(filename)

	recordNumbers, found := getRecordNumbers(fh, directory, user, name, extension, diskParams)

	if found {
		dumpRecords(fh, recordNumbers)
	} else {
		fmt.Println("File not found")
	}

	fmt.Println()
	fmt.Println()
}

func exportCommand(fh *os.File, directory []byte, filename string, exportDirectory string, diskParams utils.DiskParams) {
	user, name, extension := splitFilename(filename)

	recordNumbers, found := getRecordNumbers(fh, directory, user, name, extension, diskParams)

	if found {
		exportRecords(fh, recordNumbers, filename, exportDirectory)
	} else {
		fmt.Println("File not found")
	}

	fmt.Println()
}

func readDirectory(fh *os.File, diskParams utils.DiskParams) ([]byte, error) {
	indexes := []int{}

	if diskParams.Type == 0 {
		// read sector 30 and 34 (the directory on an H-17 SSSD disk)
		indexes = []int{30, 34}
	}

	if diskParams.Type == 1 {
		// read sectors 30, 33, 36, and 39 (the directory on an H-37 SSSD disk)
		indexes = []int{30, 33, 36, 39}
	}

	directory, err := utils.ReadSectors(fh, indexes)

	return directory, err
}

func Export(fh *os.File, exportSpec string, exportDirectory string, diskParams utils.DiskParams) {
	directory, err := readDirectory(fh, diskParams)
	if err != nil {
		fmt.Println(err.Error)
		return
	}

	exportCommand(fh, directory, exportSpec, exportDirectory, diskParams)
}

func Cat(fh *os.File, diskParams utils.DiskParams) {
	directory, err := readDirectory(fh, diskParams)
	if err != nil {
		fmt.Println(err.Error)
		return
	}

	dirCommand(fh, directory, diskParams)
}

func Menu(reader *bufio.Reader, fh *os.File, exportDirectory string, diskParams utils.DiskParams) {
	directory, err := readDirectory(fh, diskParams)
	if err != nil {
		fmt.Println(err.Error)
		return
	}

	// prompt for command and process it
	done := false
	for !done {
		// display prompt and read command
		fmt.Printf("CP/M> ")
		line, err := reader.ReadString('\n')
		utils.CheckAndExit(err)

		// process the command
		line = strings.TrimSpace(line)
		parts := strings.Split(line, " ")

		if parts[0] == "exit" {
			fmt.Println()
			done = true
		} else if parts[0] == "stats" {
			fmt.Println("not implemented")
		} else if parts[0] == "cat" {
			catCommand(fh, directory, false, diskParams)
		} else if parts[0] == "cats" {
			catCommand(fh, directory, true, diskParams)
		} else if parts[0] == "dir" {
			dirCommand(fh, directory, diskParams)
		} else if parts[0] == "type" {
			typeCommand(fh, directory, parts[1], diskParams)
		} else if parts[0] == "dump" {
			dumpCommand(fh, directory, parts[1], diskParams)
		} else if parts[0] == "export" {
			exportCommand(fh, directory, parts[1], exportDirectory, diskParams)
		} else {
			help()
			fmt.Println()
		}
	}
}

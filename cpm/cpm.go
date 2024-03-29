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

func sectorsToRecords(sectorIndexes []int, recordsPerSector int) []int {
	recordIndexes := []int{}

	for _, sectorIndex := range sectorIndexes {
		recordIndex := sectorIndex * recordsPerSector

		for i := 0; i < recordsPerSector; i++ {
			recordIndexes = append(recordIndexes, recordIndex)

			recordIndex += 1
		}
	}

	return recordIndexes
}

func blockToSectors(block int, sectorsPerBlock int, sectorMap [][]int, blocksPerMap, dirBase int) []int {
	index := block % blocksPerMap
	track := (block / blocksPerMap) * sectorsPerBlock
	offsetsInBlock := sectorMap[index]

	sectorIndexes := []int{}
	for _, offsetInBlock := range offsetsInBlock {
		sectorIndex := dirBase + track + offsetInBlock

		sectorIndexes = append(sectorIndexes, sectorIndex)
	}

	return sectorIndexes
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

// return all record numbers for a file
func allRecords(blocks []int, recordCount int, diskGeometry utils.DiskGeometry, diskType utils.DiskType) []int {
	sectorsPerBlock := 20
	directoryFirstSector := 30
	recordsPerSector := 2
	blocksPerMap := 5
	sectorMap := [][]int{}

	if diskType == utils.H37 {
		sectorMap = [][]int{
			{0, 3, 6, 9},
			{2, 5, 8, 1},
			{4, 7, 10, 13},
			{16, 19, 12, 15},
			{18, 11, 14, 17},
		}
	}

	if diskType == utils.H17 {
		sectorMap = [][]int{
			{0, 4, 8, 2},
			{6, 1, 5, 9},
			{3, 7, 10, 14},
			{18, 12, 16, 11},
			{15, 19, 13, 17},
		}
	}

	records := []int{}

	for _, block := range blocks {
		blockSectors := blockToSectors(block, sectorsPerBlock, sectorMap, blocksPerMap, directoryFirstSector)
		blockRecords := sectorsToRecords(blockSectors, recordsPerSector)
		records = append(records, blockRecords...)
	}

	if recordCount > 0 {
		records = records[:recordCount]
	}

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
func catCommand(data []byte, directory []byte, details bool, diskGeometry utils.DiskGeometry, diskType utils.DiskType) {
	fmt.Println("User Name          Extent Flags         Records Blocks")

	index := 0
	entrySize := 32

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
				recordNumbers := allRecords(blocks, recordCount, diskGeometry, diskType)

				recordText := recordsToText(recordNumbers)
				fmt.Println(recordText)
			}
		}

		fmt.Println()

		index += entrySize
	}

	fmt.Println()
}

// print file-oriented directory (one line per file, not per entry)
func dirCommand(data []byte, directory []byte, diskGeometry utils.DiskGeometry, diskType utils.DiskType) {
	// for each user (0 to 31)
	for user := 0; user < 32; user++ {
		// get list of all file names with no repeats (strip flags)
		index := 0
		entrySize := 32

		fileNames := []string{}
		fileBlocks := map[string]int{}
		fileFlags := map[string]string{}

		print_user_header := true

		for index < len(directory) {
			end := index + entrySize
			entry := DirectoryEntry{}
			entry.Init(directory[index:end])

			entryUser := int(entry.User)

			if entryUser == user {
				if print_user_header {
					fmt.Println()
					fmt.Printf("User: %d\n", entryUser)
					fmt.Println("Name          Flags      Records")

					print_user_header = false
				}

				// get filename
				filename := entry.nameToText()

				if _, ok := fileBlocks[filename]; ok {
				} else {
					fileNames = append(fileNames, filename)
				}

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
				recordNumbers := allRecords(blocks, recordCount, diskGeometry, diskType)
				fileBlocks[filename] += len(recordNumbers)
			}

			index += entrySize
		}

		// for each file, print info
		for _, filename := range fileNames {
			flags := fileFlags[filename]
			size := fileBlocks[filename]
			fmt.Printf("%-12s  %s %5d\n", filename, flags, size)
		}
	}

	fmt.Println()
}

func readRecord(data []byte, recordNumber int) ([]byte) {
	sectorNumber := recordNumber / 2
	offset := recordNumber % 2

	recordBytes := []byte{}
	start1 := sectorNumber * 256
	end1 := start1 + 256
	sectorBytes := data[start1:end1]

	start := 0 + 128*offset
	end := start + 128
	recordBytes = sectorBytes[start:end]

	return recordBytes
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

func displayRecords(data []byte, recordNumbers []int) {
	// for each record in block
	for _, record := range recordNumbers {
		// read data
		recordBytes := readRecord(data, record)

		// print data
		displayText(recordBytes)
	}
}

func dumpRecords(data []byte, format string, recordNumbers []int) {
	// for each record in block
	for i, record := range recordNumbers {
		fmt.Printf("RECORD: %d\n", i)
		// read data
		recordBytes := readRecord(data, record)

		// print data
		utils.Dump(recordBytes, i, format)
		fmt.Println()
	}
}

func exportRecords(data []byte, recordNumbers []int, filename string, exportDirectory string) {
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
		recordBytes := readRecord(data, record)

		if err != nil {
			fmt.Println("Could not read record")
		} else {
			// print data
			f.Write(recordBytes)
		}
	}

	fmt.Println("Done")
}

func getRecordNumbers(data []byte, directory []byte, user int, name string, extension string, diskGeometry utils.DiskGeometry, diskType utils.DiskType) ([]int, bool) {
	recordNumbers := []int{}

	entrySize := 32

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

				blockRecordNumbers := allRecords(blocks, recordCount, diskGeometry, diskType)
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

func typeCommand(data []byte, directory []byte, filename string, diskGeometry utils.DiskGeometry, diskType utils.DiskType) {
	user, name, extension := splitFilename(filename)

	recordNumbers, found := getRecordNumbers(data, directory, user, name, extension, diskGeometry, diskType)

	if found {
		displayRecords(data, recordNumbers)
	} else {
		fmt.Println("File not found")
	}

	fmt.Println()
	fmt.Println()
}

func dumpCommand(data []byte, directory []byte, filename string, format string, diskGeometry utils.DiskGeometry, diskType utils.DiskType) {
	user, name, extension := splitFilename(filename)

	recordNumbers, found := getRecordNumbers(data, directory, user, name, extension, diskGeometry, diskType)

	if found {
		dumpRecords(data, format, recordNumbers)
	} else {
		fmt.Println("File not found")
	}

	fmt.Println()
	fmt.Println()
}

func exportCommand(data []byte, directory []byte, filename string, exportDirectory string, diskGeometry utils.DiskGeometry, diskType utils.DiskType) {
	user, name, extension := splitFilename(filename)

	recordNumbers, found := getRecordNumbers(data, directory, user, name, extension, diskGeometry, diskType)

	if found {
		exportRecords(data, recordNumbers, filename, exportDirectory)
	} else {
		fmt.Println("File not found")
	}

	fmt.Println()
}

func readDirectory(data []byte, diskGeometry utils.DiskGeometry, diskType utils.DiskType) ([]byte) {
	blocks := []int{0, 1}

	recordCount := -1
	recordNumbers := allRecords(blocks, recordCount, diskGeometry, diskType)

	directory := make([]byte, 0)
	// for each record in block
	for _, record := range recordNumbers {

		// read data
		recordBytes := readRecord(data, record)
		directory = append(directory, recordBytes...)
	}

	return directory
}

func Export(data []byte, exportSpec string, exportDirectory string, diskGeometry utils.DiskGeometry, diskType utils.DiskType) {
	directory := readDirectory(data, diskGeometry, diskType)

	exportCommand(data, directory, exportSpec, exportDirectory, diskGeometry, diskType)
}

func Cat(data []byte, diskGeometry utils.DiskGeometry, diskType utils.DiskType) {
	directory := readDirectory(data, diskGeometry, diskType)

	dirCommand(data, directory, diskGeometry, diskType)
}

func Menu(reader *bufio.Reader, data []byte, exportDirectory string, diskGeometry utils.DiskGeometry, diskType utils.DiskType) {
	directory := readDirectory(data, diskGeometry, diskType)
	dump_format := "octal"

	// prompt for command and process it
	done := false
	for !done {
		// display prompt and read command
		fmt.Printf("CP/M> ")
		line, err := reader.ReadString('\n')
		utils.CheckAndExit(err)

		// process the command
		line = strings.TrimSpace(line)
		utils.EchoInput(line)
		parts := strings.Split(line, " ")

		if parts[0] == "exit" {
			fmt.Println()
			done = true
		} else if parts[0] == "stats" {
			fmt.Println("not implemented")
		} else if parts[0] == "cat" {
			catCommand(data, directory, false, diskGeometry, diskType)
		} else if parts[0] == "cats" {
			catCommand(data, directory, true, diskGeometry, diskType)
		} else if parts[0] == "dir" {
			dirCommand(data, directory, diskGeometry, diskType)
		} else if parts[0] == "type" {
			if len(parts) > 1 {
				typeCommand(data, directory, parts[1], diskGeometry, diskType)
			} else {
				fmt.Println("File name required")
			}
		} else if parts[0] == "dump" {
			if len(parts) > 1 {
				format := dump_format
				if len(parts) > 2 {
					format = parts[2]
				}
				dumpCommand(data, directory, parts[1], format, diskGeometry, diskType)
			} else {
				fmt.Println("File name required")
			}
		} else if parts[0] == "export" {
			if len(parts) > 1 {
				exportCommand(data, directory, parts[1], exportDirectory, diskGeometry, diskType)
			} else {
				fmt.Println("File name required")
			}
		} else {
			help()
			fmt.Println()
		}
	}
}

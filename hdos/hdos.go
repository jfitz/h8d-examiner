/*
Package hdos of H-8/H-89 disk reader
*/
package hdos

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/jfitz/h8d-examiner/utils"
	"os"
	"strings"
)

func help() {
	fmt.Println("stats  - display statistics")
	fmt.Println("cat    - list files on disk")
	fmt.Println("dir    - same as CAT")
	fmt.Println("type   - display contents of file")
	fmt.Println("dump   - dump contents of file")
	fmt.Println("export - copy file to your filesystem")
	fmt.Println("exit   - exit to main level")
}

func readSectorPair(fh *os.File, sectorIndex int) ([]byte, error) {
	// read 2 sectors (512 bytes)
	first, err := utils.ReadSector(fh, sectorIndex)
	if err != nil {
		return []byte{}, errors.New("Cannot read first directory sector")
	}

	second, err := utils.ReadSector(fh, sectorIndex+1)
	if err != nil {
		return []byte{}, errors.New("Cannot read second directory sector")
	}

	directoryBlock := append(first, second...)

	return directoryBlock, nil
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

func printDirectoryBlock(directoryBlock []byte, grtSector []byte, sectorsPerGroup int, details bool) {
	// parse and print 22 entries of 23 bytes each
	for i := 0; i < 22; i++ {
		start := i * 23
		end := start + 23
		entry := directoryBlock[start:end]
		nameBytes := entry[0:8]
		if nameBytes[0] < 0xfe {
			extensionBytes := entry[8:11]
			name := string(utils.TrimSlice(nameBytes))
			extension := string(utils.TrimSlice(extensionBytes))
			project := entry[12]
			version := entry[13]
			flagByte := entry[14]
			flags := flagsToText(flagByte)

			modifyDateBytes := entry[21:23]
			modifyDate := dateToText(modifyDateBytes)

			firstCluster := entry[16]
			lastCluster := entry[17]
			lastSector := int(entry[18])
			usedSectors := getSectors(grtSector, firstCluster, lastCluster, lastSector, sectorsPerGroup)
			usedSectorCount := len(usedSectors)

			if details {
				createDateBytes := entry[19:21]
				createDate := dateToText(createDateBytes)

				allocSectors := getSectors(grtSector, firstCluster, lastCluster, sectorsPerGroup, sectorsPerGroup)
				allocSectorCount := len(allocSectors)

				fmt.Printf("%-8s.%-3s[%04d];%03d    %s     %s    %s   %4d   %4d\n", name, extension, project, version, flags, createDate, modifyDate, usedSectorCount, allocSectorCount)
			} else {
				fmt.Printf("%-8s.%-3s    %s     %s   %4d\n", name, extension, flags, modifyDate, usedSectorCount)
			}
		}
	}
}

func getFileSectors(wantedName string, directoryBlock []byte, grtSector []byte, sectorsPerGroup int) ([]int, bool) {
	// parse and print 22 entries of 23 bytes each
	for i := 0; i < 22; i++ {
		start := i * 23
		end := start + 23
		entry := directoryBlock[start:end]
		nameBytes := entry[0:8]
		if nameBytes[0] < 0xfe {
			extensionBytes := entry[8:11]
			name := string(utils.TrimSlice(nameBytes))
			extension := string(utils.TrimSlice(extensionBytes))

			firstCluster := entry[16]
			lastCluster := entry[17]
			lastSector := int(entry[18])
			usedSectors := getSectors(grtSector, firstCluster, lastCluster, lastSector, sectorsPerGroup)

			filename := name + "." + extension
			if filename == wantedName {
				return usedSectors, true
			}
		}
	}

	return []int{}, false
}

type Label struct {
	Dis  int
	Grt  int
	Spg  int
	Ver  int
	Siz  int
	Pss  int
	Spt  int
	Text string
}

func (label *Label) Init(sector []byte) {
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
	if label.Ver >= 0x20 {
		// extract and validate number of sectors
		label.Siz = int(sector[12]) + int(sector[13])*256

		// extract and validate sector size == 256
		label.Pss = int(sector[14]) + int(sector[15])*256

		// extract and validate flags 0 => 40tk1s 1=> 40tk2s 2=> 80tk1s 3=> 80tk2s

		// extract and validate sectors per track == 10
		label.Spt = int(sector[79])
	}

	// extract and validate label ASCII text, zero terminated
	labelBytes := utils.TrimSlice(sector[17:77])

	label.Text = string(labelBytes)

	// if version 20h: num sectors match flags 0 => 400 1 => 800 2 => 800 3 => 1600
}

func (label Label) Print() {
	fmt.Printf("First directory sector: 0x%02X (%d)\n", label.Dis, label.Dis)
	fmt.Printf("GRT sector: 0x%02X (%d)\n", label.Grt, label.Grt)
	fmt.Printf("Sectors per group: %d\n", label.Spg)
	fmt.Printf("INIT.ABS version: 0x%02X\n", label.Ver)
	fmt.Printf("Number of sectors: %d\n", label.Siz)
	fmt.Printf("Sector size: %d\n", label.Pss)
	fmt.Printf("Sectors per track: %d\n", label.Spt)
	fmt.Printf("Label: %s\n", label.Text)
}

func catCommand(fh *os.File, label Label, grtSector []byte) {
	fmt.Println("Name                      Flags    Created        Modified      Used  Allocated")

	// start with first directory sector
	sectorIndex := label.Dis

	for sectorIndex != 0 {
		directoryBlock, err := readSectorPair(fh, sectorIndex)
		if err != nil {
			fmt.Println(err.Error())
		}

		printDirectoryBlock(directoryBlock, grtSector, label.Spg, true)

		// read 6 bytes
		vectorBytes := directoryBlock[506:512]

		// bytes [4] and [5] are index of next directory pair
		sectorIndex = int(vectorBytes[4]) + int(vectorBytes[5])*256
	}

	fmt.Println()
}

func dirCommand(fh *os.File, label Label, grtSector []byte) {
	fmt.Println("Name            Flags    Modified      Used")

	// start with first directory sector
	sectorIndex := label.Dis

	for sectorIndex != 0 {
		directoryBlock, err := readSectorPair(fh, sectorIndex)
		if err != nil {
			fmt.Println(err.Error())
		}

		printDirectoryBlock(directoryBlock, grtSector, label.Spg, false)

		// read 6 bytes
		vectorBytes := directoryBlock[506:512]

		// bytes [4] and [5] are index of next directory pair
		sectorIndex = int(vectorBytes[4]) + int(vectorBytes[5])*256
	}

	fmt.Println()
}

func fileSectors(fh *os.File, label Label, grtSector []byte, wantedFilename string) ([]int, bool) {
	// start with first directory sector
	sectorIndex := label.Dis

	fileSectors := []int{}

	found := false

	for sectorIndex != 0 && !found {
		directoryBlock, err := readSectorPair(fh, sectorIndex)
		if err != nil {
			fmt.Println(err.Error())
		}

		fileSectors, found = getFileSectors(wantedFilename, directoryBlock, grtSector, label.Spg)

		// read 6 bytes
		vectorBytes := directoryBlock[506:512]

		// bytes [4] and [5] are index of next directory pair
		sectorIndex = int(vectorBytes[4]) + int(vectorBytes[5])*256
	}

	return fileSectors, found
}

func typeCommand(fh *os.File, label Label, grtSector []byte, filename string) {
	sectorNumbers, found := fileSectors(fh, label, grtSector, filename)

	if found {
		// for each sector
		for _, sectorNumber := range sectorNumbers {
			sectorBytes, err := utils.ReadSector(fh, sectorNumber)
			if err != nil {
				fmt.Println("Count not read sector")
			} else {
				text := string(sectorBytes)
				fmt.Print(text)
			}
		}

		fmt.Println()
		fmt.Println()
	} else {
		fmt.Println("File not found")
	}

	fmt.Println()
}

func dumpCommand(fh *os.File, label Label, grtSector []byte, filename string, format string) {
	sectorNumbers, found := fileSectors(fh, label, grtSector, filename)

	if found {
		fmt.Println()

		// for each sector
		for i, sectorNumber := range sectorNumbers {
			sectorBytes, err := utils.ReadSector(fh, sectorNumber)
			if err != nil {
				fmt.Println("Count not read sector")
			} else {
				utils.Dump(sectorBytes, i, format)
				fmt.Println()
			}
		}

		fmt.Println()
		fmt.Println()
	} else {
		fmt.Println("File not found")
	}

	fmt.Println()
}

func exportCommand(fh *os.File, label Label, grtSector []byte, filename string, exportDirectory string) {
	sectorNumbers, found := fileSectors(fh, label, grtSector, filename)

	if found {
		fmt.Println("Exporting file...")

		// open file
		exportFilename := exportDirectory + "/" + filename
		f, err := os.Create(exportFilename)
		defer f.Close()

		if err != nil {
			fmt.Println("Cannot open file")
			return
		}

		// for each sector
		for _, sectorNumber := range sectorNumbers {
			sectorBytes, err := utils.ReadSector(fh, sectorNumber)
			if err != nil {
				fmt.Println("Count not read sector")
			} else {
				// write sector
				f.Write(sectorBytes)
			}
		}

		fmt.Println("Done")
	} else {
		fmt.Println("File not found")
	}

	fmt.Println()
}

func readLabel(fh *os.File) (Label, error) {
	label := Label{}

	// read sector 9
	sectorIndex := 9
	sector, err := utils.ReadSector(fh, sectorIndex)
	if err != nil {
		return label, errors.New("Cannot read sector 9")
	}

	label.Init(sector)

	return label, nil
}

func Export(fh *os.File, exportSpec string, exportDirectory string) {
	label, err := readLabel(fh)
	if err != nil {
		fmt.Println(err.Error)
		return
	}

	// read Group Reservation Table (GRT)
	grtSector, err := utils.ReadSector(fh, label.Grt)
	utils.CheckAndExit(err)

	exportCommand(fh, label, grtSector, exportSpec, exportDirectory)
}

func Cat(fh *os.File) {
	label, err := readLabel(fh)
	if err != nil {
		fmt.Println(err.Error)
		return
	}

	// read Group Reservation Table (GRT)
	grtSector, err := utils.ReadSector(fh, label.Grt)
	utils.CheckAndExit(err)

	dirCommand(fh, label, grtSector)
}

func Menu(reader *bufio.Reader, fh *os.File, exportDirectory string) {
	label, err := readLabel(fh)
	if err != nil {
		fmt.Println(err.Error)
		return
	}

	// check text label
	labelError := false
	for _, c := range label.Text {
		if c < 32 || c > 126 {
			labelError = true
		}
	}

	if labelError {
		fmt.Println("This disk has a strange label")
	}

	// read Group Reservation Table (GRT)
	grtSector, err := utils.ReadSector(fh, label.Grt)
	utils.CheckAndExit(err)

	dump_format := "octal"

	// prompt for command and process it
	done := false
	for !done {
		// display prompt and read command
		fmt.Printf("HDOS> ")
		line, err := reader.ReadString('\n')
		utils.CheckAndExit(err)

		// process the command
		line = strings.TrimSpace(line)
		parts := strings.Split(line, " ")

		if parts[0] == "exit" {
			fmt.Println()
			done = true
		} else if parts[0] == "stats" {
			label.Print()

			freeSectors := getSectors(grtSector, 0, 0, label.Spg, label.Spg)
			freeSectorCount := len(freeSectors)
			fmt.Printf("Free sectors: %d\n", freeSectorCount)
			fmt.Println()
		} else if parts[0] == "cat" {
			catCommand(fh, label, grtSector)
		} else if parts[0] == "dir" {
			dirCommand(fh, label, grtSector)
		} else if parts[0] == "type" {
			if len(parts) > 1 {
				typeCommand(fh, label, grtSector, parts[1])
			} else {
				fmt.Println("File name required")
			}
		} else if parts[0] == "dump" {
			if len(parts) > 1 {
				format := dump_format
				if len(parts) > 2 {
					format = parts[2]
				}
				dumpCommand(fh, label, grtSector, parts[1], format)
			} else {
				fmt.Println("File name required")
			}
		} else if parts[0] == "export" {
			if len(parts) > 1 {
				exportCommand(fh, label, grtSector, parts[1], exportDirectory)
			} else {
				fmt.Println("File name required")
			}
		} else {
			help()
			fmt.Println()
		}
	}
}

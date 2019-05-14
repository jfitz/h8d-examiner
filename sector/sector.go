/*
Package sector of H-8/H-89 disk reader
*/
package sector

import (
	"bufio"
	"fmt"
	"github.com/jfitz/h8d-examiner/utils"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func help() {
	fmt.Println("exit  - exit to main level")
	fmt.Println("<RET> - dump next sector")
	fmt.Println("nnn   - dump sector nnn")
	fmt.Println("octal - show dump in octal")
	fmt.Println("hex   - show dump in hex")
}

func dumpSector(fh *os.File, sectorIndex int, base string) error {
	sector, err := utils.ReadSector(fh, sectorIndex)
	if err != nil {
		return err
	}

	return utils.Dump(sector, sectorIndex, base)
}

func Menu(reader *bufio.Reader, fh *os.File) {
	// set default values
	base := "hex"
	sectorIndex := 0
	lastWasDump := false

	numberPattern, err := regexp.Compile("^\\d+$")
	utils.CheckAndExit(err)

	// display the first sector
	err = dumpSector(fh, sectorIndex, base)
	utils.CheckAndExit(err)
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
		utils.CheckAndExit(err)

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
			utils.CheckAndExit(err)
			fmt.Println()
			lastWasDump = true
		} else if numberPattern.MatchString(line) {
			sectorIndex, _ = strconv.Atoi(line)

			err = dumpSector(fh, sectorIndex, base)
			utils.CheckAndExit(err)
			fmt.Println()
			lastWasDump = true
		} else if line == "octal" {
			base = "octal"

			err = dumpSector(fh, sectorIndex, base)
			utils.CheckAndExit(err)
			fmt.Println()
			lastWasDump = true
		} else if line == "hex" {
			base = "hex"

			err = dumpSector(fh, sectorIndex, base)
			utils.CheckAndExit(err)
			fmt.Println()
			lastWasDump = true
		} else {
			help()
			fmt.Println()
		}
	}
}

/*
Package sector of H-8/H-89 disk reader
*/
package sector

import (
	"bufio"
	"fmt"
	"github.com/jfitz/h8d-examiner/utils"
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

func dumpSector(data []byte, sectorIndex int, base string) error {
	start1 := sectorIndex * 256
	end1 := start1 + 256
	sector := data[start1:end1]

	return utils.Dump(sector, sectorIndex, base)
}

func Menu(reader *bufio.Reader, data []byte) {
	// set default values
	base := "hex"
	sectorIndex := 0
	lastWasDump := false

	numberPattern, err := regexp.Compile("^\\d+$")
	utils.CheckAndExit(err)

	// display the first sector
	err = dumpSector(data, sectorIndex, base)
	utils.CheckAndExit(err)
	fmt.Println()
	lastWasDump = true

	fileSize := len(data)
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
		utils.EchoInput(line)

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

			err = dumpSector(data, sectorIndex, base)
			utils.CheckAndExit(err)
			fmt.Println()
			lastWasDump = true
		} else if numberPattern.MatchString(line) {
			sectorIndex, _ = strconv.Atoi(line)

			err = dumpSector(data, sectorIndex, base)
			utils.CheckAndExit(err)
			fmt.Println()
			lastWasDump = true
		} else if line == "octal" {
			base = "octal"

			err = dumpSector(data, sectorIndex, base)
			utils.CheckAndExit(err)
			fmt.Println()
			lastWasDump = true
		} else if line == "hex" {
			base = "hex"

			err = dumpSector(data, sectorIndex, base)
			utils.CheckAndExit(err)
			fmt.Println()
			lastWasDump = true
		} else {
			help()
			fmt.Println()
		}
	}
}

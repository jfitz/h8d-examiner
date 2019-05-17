/*
Package main of H-8/H-89 disk reader
*/
package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/jfitz/h8d-examiner/cpm"
	"github.com/jfitz/h8d-examiner/hdos"
	"github.com/jfitz/h8d-examiner/sector"
	"github.com/jfitz/h8d-examiner/utils"
	"os"
	"strings"
)

func mainHelp() {
	fmt.Println("stats - display statistics")
	fmt.Println("hdos  - interpret as HDOS disk")
	fmt.Println("cp/m  - interpret as CP/M disk")
	fmt.Println("RESETTERM - reset VT-100 terminal")
	fmt.Println("quit  - exit the program")
}

func main() {
	exportDirectoryPtr := flag.String("directory", ".", "Export to directory")
	// exportSpecPtr := flag.String("export", "*.*", "Export file specification")

	// parse command line options
	flag.Parse()

	exportDirectory := *exportDirectoryPtr
	//	exportSpec := *exportSpecPtr

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
	utils.CheckAndExit(err)

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
		utils.CheckAndExit(err)

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
			sector.Menu(reader, fh)
		} else if line == "hdos" {
			fmt.Println()
			hdos.Menu(reader, fh, exportDirectory)
		} else if line == "cp/m" {
			fmt.Println()
			cpm.Menu(reader, fh, exportDirectory)
		} else if line == "RESETTERM" {
			fmt.Println("\x1bc")
		} else {
			mainHelp()
			fmt.Println()
		}
	}
}

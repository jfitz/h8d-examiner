/*
Package utils of H-8/H-89 disk reader
*/
package utils

import (
	"bytes"
	"errors"
	"fmt"
	"os"
)

func CheckAndExit(e error) {
	if e != nil {
		fmt.Println(e.Error())
		os.Exit(1)
	}
}

func TrimSlice(slice []byte) []byte {
	n := bytes.IndexByte(slice, byte(0))

	if n > -1 {
		slice = slice[:n]
	}

	return slice
}

type DiskType int

const (
	H17 DiskType = 17
	H27 DiskType = 27
	H37 DiskType = 37
	H47 DiskType = 47
)

type DiskSides int

const (
	SingleSided DiskSides = 1
	DoubleSided DiskSides = 2
)

type DiskGeometry struct {
	Sides            DiskSides
	Tracks           int
	SectorsPerTrack  int
	BytesPerSector   int
	SectorsPerTrack0 int
}

func ReadSector(fh *os.File, sectorIndex int) ([]byte, error) {
	sector := make([]byte, 256)

	// position at the desired sector
	pos := int64(sectorIndex) * 256

	_, err := fh.Seek(pos, 0)
	if err != nil {
		return sector, errors.New("Sector does not exist")
	}

	// read the sector
	_, err = fh.Read(sector)
	if err != nil {
		return sector, err
	}

	if len(sector) != 256 {
		return sector, errors.New("Invalid sector length")
	}

	return sector, nil
}

func ReadSectors(fh *os.File, sectorIndexes []int) ([]byte, error) {
	sectors := []byte{}

	for _, index := range sectorIndexes {
		sector, err := ReadSector(fh, index)
		if err != nil {
			return sectors, err
		}

		sectors = append(sectors, sector...)
	}

	return sectors, nil
}

func dumpOctal(bytes []byte) {
	for _, b := range bytes {
		fmt.Printf(" %03o", b)
	}
}

func dumpAscii(bytes []byte) {
	for _, b := range bytes {
		if b >= ' ' && b <= 127 {
			fmt.Printf("%c", b)
		} else {
			fmt.Print(".")
		}
	}
}

func dumpSector(sector []byte, format string) {
	// print data in lines of 16 bytes
	for i := 0; i < len(sector); i += 16 {
		upper := i + 16

		if upper > len(sector) {
			upper = len(sector)
		}

		bytes := sector[i:upper]

		// print offset
		if format == "hex" {
			fmt.Printf("%02X: ", i)
		} else {
			fmt.Printf("%03o: ", i)
		}

		// print contents
		if format == "hex" {
			fmt.Printf("% 02X", bytes)
		} else {
			dumpOctal(bytes)
		}

		fmt.Print("  ")

		// print in ASCII (with dots for non-printable bytes)
		dumpAscii(bytes)

		fmt.Println()
	}
}

func Dump(sector []byte, sectorIndex int, format string) error {
	// display the sector

	// print header information
	if format == "hex" {
		fmt.Printf("Sector: %04XH (%d):\n", sectorIndex, sectorIndex)
	} else {
		highByte := sectorIndex / 256
		lowByte := sectorIndex % 256
		fmt.Printf("Sector: %03o.%03oA (%d):\n", highByte, lowByte, sectorIndex)
	}

	fmt.Println()

	dumpSector(sector, format)

	return nil
}

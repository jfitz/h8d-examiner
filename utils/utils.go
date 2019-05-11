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

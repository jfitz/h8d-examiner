/*
Dump tokenized MBASIC-80 program
*/
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

func dumpAscii(line_number int, bytes []byte, table map[int]string, table2 map[int]string, prefix byte) {
	fmt.Printf("%d ", line_number)

	for len(bytes) > 0 {
		b := bytes[0]
		b16 := int(b)

		// 0x0F - short
		if b == 0x0F {
			i := int(bytes[1])

			fmt.Printf("%d", i)

			bytes = bytes[2:]
		}

		// 0x0E - int
		if b == 0x0E {
			i := int(bytes[2])*256 + int(bytes[1])

			fmt.Printf("%d", i)

			bytes = bytes[3:]
		}

		// 0x1C - hex int
		if b == 0x1C {
			i := int(bytes[2])*256 + int(bytes[1])

			fmt.Printf("&%04X", i)

			bytes = bytes[3:]
		}

		// 0xFF - 2-byte token
		if b == 0xFF {
			code := int(bytes[1])

			if s, ok := table2[code]; ok {
				fmt.Print(s)
			} else {
				fmt.Print(".")
			}

			bytes = bytes[2:]
		}

		// 0x80 to 0xFE - 1-byte token
		if b16 >= 0x80 && b16 <= 0xFE {
			if s, ok := table[b16]; ok {
				fmt.Print(s)
			} else {
				fmt.Print(".")
			}

			bytes = bytes[1:]
		}

		// 0x32 to 0x7F - plain character
		if b >= 0x20 && b <= 0x7F {
			fmt.Printf("%c", b16)

			bytes = bytes[1:]
		}

		// 0x01 to 0x31 - 1-byte number
		if b > 0 && b < 0x20 {
			handled := false

			if b == 0x09 {
				// TAB
				fmt.Print("\\t")
				bytes = bytes[1:]
				handled = true
			}

			if b == 0x0A {
				// LF
				fmt.Print("\\n")
				bytes = bytes[1:]
				handled = true
			}

			if b == 0x0D {
				// CR
				fmt.Print("\\r")
				bytes = bytes[1:]
				handled = true
			}

			if b == 0x0E || b == 0x0F || b == 0x1C {
				// do nothing
				handled = true
			}

			if !handled {
				fmt.Printf("%d", b16)
				bytes = bytes[1:]
			}
		}

		// ignore byte of zero
		if b16 == 0 {
			bytes = bytes[1:]
		}
	}
}

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) == 0 {
		fmt.Println("No file specified")
		os.Exit(1)
	}

	// get file name
	fileName := args[0]

	bytes, _ := ioutil.ReadFile(fileName)

	// build decode tables
	table := map[int]string{}

	table[0x81] = "END"
	table[0x82] = "FOR"
	table[0x83] = "NEXT"
	table[0x84] = "DATA"
	table[0x85] = "INPUT"
	table[0x86] = "DIM"
	table[0x87] = "READ"
	table[0x88] = "LET"
	table[0x89] = "GOTO"
	table[0x8A] = "RUN"
	table[0x8B] = "IF"
	table[0x8C] = "RESTORE"
	table[0x8D] = "GOSUB"
	table[0x8E] = "RETURN"
	table[0x8F] = "REM"
	table[0x90] = "STOP"
	table[0x91] = "PRINT"
	table[0x92] = "CLEAR"
	table[0x93] = "LIST"
	table[0x94] = "NEW"
	table[0x95] = "ON"
	table[0x96] = "DEF"
	table[0x97] = "POKE"
	table[0x98] = ""
	table[0x99] = "unknown"
	table[0x9A] = "unknown"
	table[0x9B] = "LPRINT"
	table[0x9C] = "LLIST"
	table[0x9D] = "WIDTH"
	table[0x9E] = "ELSE"
	table[0x9F] = "TRACE"
	table[0xA0] = "NOTRACE"
	table[0xA1] = "SWAP"
	table[0xA2] = "ERASE"
	table[0xA3] = "EDIT"
	table[0xA4] = "ERROR"
	table[0xA5] = "RESUME"
	table[0xA6] = "DEL"
	table[0xA7] = "AUTO"
	table[0xA8] = "ERR"
	table[0xA9] = "DEFSTR"
	table[0xAA] = "POP"
	table[0xAB] = "DEFSNG"
	table[0xAC] = "DEFDBL"
	table[0xAD] = "LINE"
	table[0xAE] = "DEFINT"
	table[0xAF] = "WHILE"
	table[0xB0] = "WEND"
	table[0xB1] = "CALL"
	table[0xB2] = "WRITE"
	table[0xB3] = "COMMON"
	table[0xB4] = "CHAIN"
	table[0xB5] = "OPTION"
	table[0xB6] = "RANDOMIZE"
	table[0xB7] = "SYSTEM"
	table[0xB8] = "OPEN"
	table[0xB9] = "FIELD"
	table[0xBA] = "GET"
	table[0xBB] = "PUT"
	table[0xBC] = "CLOSE"
	table[0xBD] = "LOAD"
	table[0xBE] = "MERGE"
	table[0xBF] = "FILES"
	table[0xC0] = "NAME"
	table[0xC1] = "KILL"
	table[0xC2] = "LSET"
	table[0xC3] = "RSET"
	table[0xC4] = "SAVE"
	table[0xC5] = "RESET"
	table[0xC6] = "TEXT"
	table[0xC7] = "HOME"
	table[0xC8] = "VTAB"
	table[0xC9] = "HTAB"
	table[0xCA] = "INVERSE"
	table[0xCB] = "NORMAL"
	table[0xCC] = "GR"
	table[0xCD] = "COLOR"
	table[0xCE] = "TO"
	table[0xCF] = "THEN"
	table[0xD0] = "TAB("
	table[0xD1] = "HGR"
	table[0xD2] = "HPLOT"
	table[0xD3] = ""
	table[0xD4] = "BEEP"
	table[0xD5] = "WAIT"
	table[0xD6] = "unknown"
	table[0xD7] = "unknown"
	table[0xD8] = "unknown"
	table[0xD9] = "unknown"
	table[0xDA] = "unknown"
	table[0xDB] = " "
	table[0xDC] = "unknown"
	table[0xDD] = "HLIN"
	table[0xDE] = "VLIN"
	table[0xDF] = "PLOT"
	table[0xE0] = "STEP"
	table[0xE1] = "USR"
	table[0xE2] = "FN"
	table[0xE3] = "SPC("
	table[0xE4] = "NOT"
	table[0xE5] = "ERL"
	table[0xE6] = "RENUM"
	table[0xE7] = "STRING$"
	table[0xE8] = "USING"
	table[0xE9] = "INSTR"
	table[0xEA] = "unknown"
	table[0xEB] = "VARPTR"
	table[0xEC] = "SCRN"
	table[0xED] = "HSCRN"
	table[0xEE] = "INKEY$"
	table[0xEF] = ">"
	table[0xF0] = "="
	table[0xF1] = "<"
	table[0xF2] = "+"
	table[0xF3] = "-"
	table[0xF4] = "*"
	table[0xF5] = "/"
	table[0xF6] = "^"
	table[0xF7] = "AND"
	table[0xF8] = "OR"
	table[0xF9] = "XOR"
	table[0xFA] = "EQV"
	table[0xFB] = "IMP"
	table[0xFC] = "MOD"
	table[0xFD] = "unknown"
	table[0xFE] = "unknown"
	table[0xFF] = ""

	table2 := map[int]string{}

	table2[0x81] = "LEFT$"
	table2[0x82] = "RIGHT$"
	table2[0x83] = "MID$"
	table2[0x84] = "SGN"
	table2[0x85] = "INT"
	table2[0x86] = "unknown"
	table2[0x87] = "SQR"
	table2[0x88] = "RND"
	table2[0x89] = "SIN"
	table2[0x8A] = "LOG"
	table2[0x8B] = "EXP"
	table2[0x8C] = "COS"
	table2[0x8D] = "TAN"
	table2[0x8E] = "ATN"
	table2[0x8F] = "PEEK"
	table2[0x90] = "POS"
	table2[0x91] = "LEN"
	table2[0x92] = "STR$"
	table2[0x93] = "VAL"
	table2[0x94] = "FRE"
	table2[0x95] = "ASC"
	table2[0x96] = "CHR$"
	table2[0x97] = "SPACE$"
	table2[0x98] = "OCT$"
	table2[0x99] = "HEX$"
	table2[0x9A] = "LPOS"
	table2[0x9B] = "CINT"
	table2[0x9C] = "CSNG"
	table2[0x9D] = "CDBL"
	table2[0x9E] = "FIX"
	table2[0x9F] = "unknown"
	table2[0xA0] = "unknown"
	table2[0xA1] = "unknown"
	table2[0xA2] = "unknown"
	table2[0xA3] = "unknown"
	table2[0xA4] = "unknown"
	table2[0xA5] = "unknown"
	table2[0xA6] = "unknown"
	table2[0xA7] = "unknown"
	table2[0xA8] = "unknown"
	table2[0xA9] = "unknown"
	table2[0xAA] = "CVI"
	table2[0xAB] = "CVS"
	table2[0xAC] = "CVD"
	table2[0xAD] = "unknown"
	table2[0xAE] = "EOF"
	table2[0xAF] = "LOC"
	table2[0xB0] = "LOF"
	table2[0xB1] = "MKI$"
	table2[0xB2] = "MKS$"
	table2[0xB3] = "MKD$"
	table2[0xB4] = "VPOS"
	table2[0xB5] = "PDL"
	table2[0xB6] = "BUTTON"

	contents := bytes[1:]

	print_dump := true

	// magic address used by Microsoft BASIC-80 for Heathkit HDOS
	address := 0x7059

	// for all bytes
	for len(contents) > 0 {
		// 2 bytes for end
		end_address := int(contents[1])*256 + int(contents[0])

		// 2 bytes for line number
		line_number := int(contents[3])*256 + int(contents[2])

		// line of code
		length := end_address - address

		if length > 0 {
			payload := contents[4:length]

			if print_dump {
				// dump bytes (including NULL) as hex and ascii
				fmt.Printf("%d", length)
				fmt.Print("  ")

				fmt.Printf("%d", line_number)
				fmt.Print("  ")

				fmt.Printf("% 02X", payload)
				fmt.Println()
			}

			// print the untokenized line
			dumpAscii(line_number, payload, table, table2, 0xff)
			fmt.Println()

			if print_dump {
				fmt.Println()
			}

			contents = contents[length:]
		} else {
			contents = nil
		}

		address = end_address
	}
}

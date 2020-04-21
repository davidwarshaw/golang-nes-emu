package cpu

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
)

// PSRFlagType holds the PSR flags
type PSRFlagType struct {
	Carry     bool
	Zero      bool
	Interrupt bool
	Decimal   bool
	Break     bool
	Always1   bool
	Overflow  bool
	Negative  bool
}

type addressModeType struct {
	Accumulator string
	Branch      string
	Immediate   string
	ZeroPage    string
	ZeroPageX   string
	ZeroPageY   string
	Absolute    string
	AbsoluteX   string
	AbsoluteY   string
	Indirect    string
	IndirectX   string
	IndirectY   string
	Implied     string
	NA          string
}

// AddressMode is the string names of the address modes as found in the opcode CSV
var AddressMode = addressModeType{
	Accumulator: "Accumulator",
	Branch:      "Branch",
	Immediate:   "Immediate",
	ZeroPage:    "Zero Page",
	ZeroPageX:   "Zero Page, X",
	ZeroPageY:   "Zero Page, Y",
	Absolute:    "Absolute",
	AbsoluteX:   "Absolute, X",
	AbsoluteY:   "Absolute, Y",
	Indirect:    "Indirect",
	IndirectX:   "Indirect, X",
	IndirectY:   "Indirect, Y",
	Implied:     "Implied",
	NA:          "N/A",
}

type opcodeMeta struct {
	opcode      byte
	instruction string
	mode        string
	bytes       int
	cycles      int
}

// Opcodes is a slice of all CPU opcodes with instruction and addressing mode
var Opcodes []opcodeMeta

func init() {
	fmt.Println("Initializing CPU")
	recordFile, err := os.Open("cpu-6502-opcodes.csv")
	if err != nil {
		log.Fatal(err)
	}
	reader := csv.NewReader(recordFile)
	i := -1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if i >= 0 {
			hex64, err := strconv.ParseUint(record[0], 16, 64)
			if err != nil {
				log.Fatal(err)
			}
			opcode := byte(hex64)
			if byte(i) != opcode {
				fmt.Printf("i: %X opcode: %X\n", i, opcode)
				log.Fatal("opcode out of order")
			}

			instruction := record[1]
			// Replace empty codes with NOPs
			if len(instruction) == 0 {
				instruction = "NOP"
			}

			var bytes int
			if len(record[3]) > 0 {
				bytes, err = strconv.Atoi(record[3])
				if err != nil {
					log.Fatal(err)
				}
			} else {
				bytes = 0
			}

			var cycles int
			if len(record[4]) > 0 {
				cycles, err = strconv.Atoi(record[4])
				if err != nil {
					log.Fatal(err)
				}
			} else {
				cycles = 0
			}

			Opcodes = append(Opcodes, opcodeMeta{opcode, instruction, record[2], bytes, cycles})
		}
		i++
	}

	fmt.Printf("%d opcodes\n", len(Opcodes))
}

// OpcodeDump outputs all OpCodes
func OpcodeDump() {
	for index, element := range Opcodes {
		fmt.Printf("index: %X hex: %X opcode: %v\n", index, element.opcode, element)
	}
}

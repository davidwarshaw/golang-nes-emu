package cpu

import (
	"fmt"
)

func fmtForMode(mode string) string {
	switch mode {
	case AddressMode.Immediate:
		return "#$%2.2X"
	//
	case AddressMode.ZeroPage, AddressMode.Branch:
		return "$%2.2X"
	case AddressMode.ZeroPageX:
		return "$%2.2X, X"
	case AddressMode.ZeroPageY:
		return "$%2.2X, Y"
	//
	case AddressMode.Absolute:
		return "$%2.2X%2.2X"
	case AddressMode.AbsoluteX:
		return "$%2.2X%2.2X, X"
	case AddressMode.AbsoluteY:
		return "$%2.2X%2.2X, Y"
	//
	case AddressMode.Indirect:
		return "($%2.2X%2.2X)"
	case AddressMode.IndirectX:
		return "($%2.2X, X)"
	case AddressMode.IndirectY:
		return "($%2.2X), Y"
	//
	default:
		return ""
	}
}

func readHeader(buffer []byte) {
	header := buffer[:16]
	for i, b := range header {
		fmt.Printf("%d %X\n", i, b)
	}
}

// Disassemble takes an array of byte code and outputs assembler instructions
func Disassemble(buffer []byte) {
	pc := 0
	for pc < len(buffer) {
		// Pop a byte off the stack
		var opcodeByte byte
		opcodeByte = buffer[pc]
		pc++
		opcodeMeta := Opcodes[opcodeByte]

		format := fmtForMode(opcodeMeta.mode)

		fmt.Printf("%s ", opcodeMeta.instruction)
		var operand []byte
		switch opcodeMeta.bytes {
		case 2:
			operand = append(operand, buffer[pc])
			pc++
			fmt.Printf(format, operand[0])
		case 3:
			operand = append(operand, buffer[pc])
			pc++
			operand = append(operand, buffer[pc])
			pc++
			// Little endian, so reverse the bytes
			fmt.Printf(format, operand[1], operand[0])
		}
		fmt.Println()
	}
}

package cpu

import (
	"fmt"
	"log"
)

type registersType struct {
	a  byte
	x  byte
	y  byte
	ps PSRFlagType
	pc uint16
	sp byte
}

var register registersType

var memory [0xFFFF]byte

const ramOffset = 0x0000
const stackPointerZero = 0x01ff // NOTE: stack grows down!
const ppuOffset = 0x2000
const apuOffset = 0x4000
const controllerOneOffset = 0x4016
const controllerTwoOffset = 0x4016
const romOffset = 0x8000
const isrVector = 0xFFFE

var cycle int

func sumWillOverflow(a byte, b byte) bool {
	return a > 0 && b > 0xFF-a
}

func willCarry(a byte, b byte) bool {
	return false
}

func willBorrow(a byte, b byte) bool {
	return false
}

func littleEndianCompose(low byte, high byte) uint16 {
	return uint16(high)<<8 + uint16(low)
}

func lowByteOfWord(word uint16) byte {
	return byte(word)
}

func highByteOfWord(word uint16) byte {
	return byte(word >> 8)
}

func flagsFromByte(flagsByte byte) {
	register.ps.Carry = ((flagsByte << 7) >> 7) == 1
	register.ps.Zero = ((flagsByte << 6) >> 7) == 1
	register.ps.Interrupt = ((flagsByte << 5) >> 7) == 1
	register.ps.Decimal = ((flagsByte << 4) >> 7) == 1
	register.ps.Break = ((flagsByte << 3) >> 7) == 1
	register.ps.Always1 = ((flagsByte << 2) >> 7) == 1
	register.ps.Overflow = ((flagsByte << 1) >> 7) == 1
	register.ps.Negative = ((flagsByte << 0) >> 7) == 1
}

func byteFromFlags() byte {
	var flagsByte byte
	if register.ps.Carry {
		flagsByte = flagsByte | 0b00000001
	}
	if register.ps.Zero {
		flagsByte = flagsByte | 0b00000010
	}
	if register.ps.Interrupt {
		flagsByte = flagsByte | 0b00000100
	}
	if register.ps.Decimal {
		flagsByte = flagsByte | 0b00001000
	}
	if register.ps.Break {
		flagsByte = flagsByte | 0b00010000
	}
	if register.ps.Always1 {
		flagsByte = flagsByte | 0b00100000
	}
	if register.ps.Overflow {
		flagsByte = flagsByte | 0b01000000
	}
	if register.ps.Negative {
		flagsByte = flagsByte | 0b10000000
	}
	return flagsByte
}

func setFlagsForResult(result byte) byte {
	register.ps.Negative = result > 0x7F
	register.ps.Zero = result == 0
	return result
}

func addressFromMode(mode string, operands []byte) uint16 {
	// NOTE: Immediate mode is not represented here, because there is no address to resolve
	switch mode {
	case AddressMode.ZeroPage:
		return uint16(operands[0])
	case AddressMode.ZeroPageX:
		return uint16(operands[0] + register.x)
	case AddressMode.ZeroPageY:
		return uint16(operands[0] + register.y)
	//
	case AddressMode.Absolute:
		return littleEndianCompose(operands[0], operands[1])
	case AddressMode.AbsoluteX:
		return littleEndianCompose(operands[0], operands[1]) + uint16(register.x)
	case AddressMode.AbsoluteY:
		return littleEndianCompose(operands[0], operands[1]) + uint16(register.y)
	//
	case AddressMode.Indirect:
		address := littleEndianCompose(operands[0], operands[1])
		pointerLow, pointerHigh := memory[address], memory[address+1]
		return littleEndianCompose(pointerLow, pointerHigh)
	case AddressMode.IndirectX:
		address := uint16(operands[0] + register.x)
		pointerLow, pointerHigh := memory[address], memory[address+1]
		return littleEndianCompose(pointerLow, pointerHigh)
	case AddressMode.IndirectY:
		address := uint16(operands[0])
		pointerLow, pointerHigh := memory[address], memory[address+1]
		return littleEndianCompose(pointerLow, pointerHigh) + uint16(register.y)
	//
	default:
		return 0
	}
}

func nop() {
}

func push(value byte) {
	memory[register.sp] = value
	// NOTE: stack grows down!
	register.sp--
}

func pop() byte {
	value := memory[register.sp]
	// NOTE: stack grows down!
	register.sp++
	return value
}

func processOpcode(opcodeMeta opcodeMeta, operands []byte) {
	var extraCycles int
	var m *byte
	if opcodeMeta.mode == AddressMode.Accumulator {
		m = &register.a
	} else if opcodeMeta.mode == AddressMode.Immediate {
		m = &operands[0]
	} else if opcodeMeta.mode == AddressMode.Branch {
		m = &operands[0]
	} else {
		m = &memory[addressFromMode(opcodeMeta.mode, operands)]
	}
	switch opcodeMeta.instruction {
	// Storage
	case "LDA":
		register.a = setFlagsForResult(*m)
	case "LDX":
		register.a = setFlagsForResult(*m)
	case "LDY":
		register.y = setFlagsForResult(*m)
	case "STA":
		*m = register.a
	case "STX":
		*m = register.x
	case "STY":
		*m = register.y
	case "TAX":
		register.x = setFlagsForResult(register.a)
	case "TAY":
		register.y = setFlagsForResult(register.a)
	case "TXS":
		memory[register.sp] = register.x
	case "TSX":
		register.x = memory[register.sp]
	case "TXA":
		register.a = setFlagsForResult(register.x)
	case "TYA":
		register.a = setFlagsForResult(register.y)
	// Math
	case "ADC":
		register.ps.Carry = willCarry(register.a, *m)
		register.a = register.a + *m
	case "DEC":
		*m = setFlagsForResult(*m - 1)
	case "DEX":
		register.x = setFlagsForResult(register.x - 1)
	case "DEY":
		register.y = setFlagsForResult(register.y - 1)
	case "INC":
		*m = setFlagsForResult(*m + 1)
	case "INX":
		register.x = setFlagsForResult(register.x + 1)
	case "INY":
		register.y = setFlagsForResult(register.y + 1)
	case "SBC":
		register.ps.Carry = willBorrow(register.a, *m)
		register.a = register.a - *m
	// Bitwise
	case "AND":
		register.a = setFlagsForResult(register.a & *m)
	case "ASL":
		// Set the carry bit equal to the 7th bit
		register.ps.Carry = (*m >> 7) == 1
		*m = *m << 1
	case "BIT":
		register.ps.Zero = (register.a & *m) == 0
	case "EOR":
		register.a = setFlagsForResult(register.a ^ *m)
	case "LSR":
		// Set the carry bit equal to the 0th bit
		register.ps.Carry = (*m << 7) != 0
		*m = *m >> 1
	case "ORA":
		register.a = setFlagsForResult(register.a | *m)
	case "ROL":
		var carryBit byte
		if register.ps.Carry {
			carryBit = 1
		}
		// Set the carry bit equal to the 7th bit
		register.ps.Carry = (*m >> 7) == 1
		*m = *m << 1
		// Put the old carry bit into the 0th bit
		*m = *m | byte(carryBit)
	case "ROR":
		var carryBit byte
		if register.ps.Carry {
			carryBit = 1
		}
		// Set the carry bit equal to the 0th bit
		register.ps.Carry = (*m << 7) != 0
		*m = *m >> 1
		// Put the old carry bit into the 7th bit
		*m = *m | byte(carryBit<<7)
	// Branch
	case "BCS":
		if register.ps.Carry {
			register.pc += uint16(*m)
			extraCycles++
		}
	case "BCC":
		if !register.ps.Carry {
			register.pc += uint16(*m)
			extraCycles++
		}
	case "BEQ":
		if register.ps.Zero {
			register.pc += uint16(*m)
			extraCycles++
		}
	case "BNE":
		if !register.ps.Zero {
			register.pc += uint16(*m)
			extraCycles++
		}
	case "BMI":
		if register.ps.Negative {
			register.pc += uint16(*m)
			extraCycles++
		}
	case "BPL":
		if !register.ps.Negative {
			register.pc += uint16(*m)
			extraCycles++
		}
	case "BVS":
		if register.ps.Overflow {
			register.pc += uint16(*m)
			extraCycles++
		}
	case "BVC":
		if !register.ps.Overflow {
			register.pc += uint16(*m)
			extraCycles++
		}
	// Jump
	case "JMP":
		// Redo the address
		newPc := addressFromMode(opcodeMeta.mode, operands)
		register.pc = newPc
	case "JSR":
		push(highByteOfWord(register.pc))
		push(lowByteOfWord(register.pc))
		// Redo the address
		newPc := addressFromMode(opcodeMeta.mode, operands)
		register.pc = newPc
	case "RTI":
		flagsFromByte(pop())
		low := pop()
		high := pop()
		register.pc = littleEndianCompose(low, high)
	case "RTS":
		low := pop()
		high := pop()
		register.pc = littleEndianCompose(low, high) + 1
	// Registers
	case "CLC":
		register.ps.Carry = false
	case "SEC":
		register.ps.Carry = true
	case "CLD":
		register.ps.Decimal = false
	case "SED":
		register.ps.Decimal = true
	case "CLI":
		register.ps.Interrupt = false
	case "SEI":
		register.ps.Interrupt = true
	case "CLV":
		register.ps.Overflow = false
	// Compare
	case "CMP":
		register.ps.Carry = register.a >= *m
		setFlagsForResult(register.a)
	case "CPX":
		register.ps.Carry = register.x >= *m
		setFlagsForResult(register.x)
	case "CPY":
		register.ps.Carry = register.y >= *m
		setFlagsForResult(register.y)
	// Stack
	case "PHA":
		push(register.a)
	case "PLA":
		register.a = pop()
	case "PHP":
		push(byteFromFlags())
	case "PLP":
		flagsFromByte(pop())
	// System
	case "BRK":
		register.pc++
	case "NOP":
		// Do nothing
	default:
		log.Fatal("unrecognized opcode")
	}

	cycle += opcodeMeta.cycles + extraCycles
}

func execute(buffer []byte) {
	fmt.Printf("Registers: %v\n", register)
	for {
		register.pc++
		opcode := buffer[register.pc]

		opcodeMeta := Opcodes[opcode]

		var operand []byte
		switch opcodeMeta.bytes {
		case 2:
			register.pc++
			operand = append(operand, buffer[register.pc])
		case 3:
			register.pc++
			operand = append(operand, buffer[register.pc])

			register.pc++
			operand = append(operand, buffer[register.pc])
		}

		processOpcode(opcodeMeta, operand)
	}
}

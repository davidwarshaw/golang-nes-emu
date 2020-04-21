package main

import (
	"fmt"
	"log"
	"nes-go/internal/cpu"
	"nes-go/internal/rom"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("expected ROM filename")
	}
	command := os.Args[1]
	fmt.Println("Command: " + command)

	romFilePath := os.Args[2]
	fmt.Println("ROM file: " + romFilePath)

	buffer, err := rom.Read(romFilePath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d bytes\n", len(buffer))

	switch command {
	case "hexDump":
		rom.HexDump(buffer)
	case "opcodeDump":
		cpu.OpcodeDump()
	case "disassemble":
		cpu.Disassemble(buffer)
	}
}

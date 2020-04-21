package rom

import (
	"fmt"
	"io/ioutil"
)

// Read takes a path to a file and returns a byte array
func Read(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

// HexDump outputs a byte array in hexdump format
func HexDump(buffer []byte) {
	for index, element := range buffer {
		if index%16 == 0 {
			fmt.Println()
			fmt.Printf("%4.6X", index)
		}
		fmt.Printf("%4.2X", element)
	}
	fmt.Println()
}

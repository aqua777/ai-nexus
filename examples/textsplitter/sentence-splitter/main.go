package main

import (
	"fmt"
	"os"

	"github.com/aqua777/ai-flow/examples/textsplitter/sentence-splitter/funcs"
)

func main() {
	// Read the example file
	content, err := os.ReadFile("basic.txt")
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}
	funcs.TestSplitters(string(content))
}

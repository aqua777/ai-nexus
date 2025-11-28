package main

import (
	"fmt"
	"os"
	_ "github.com/aqua777/ai-flow/dotenv"
)

func main() {
	for _, env := range os.Environ() {
		fmt.Println(env)
	}
}

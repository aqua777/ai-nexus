package main

import (
	"fmt"
	"os"
	"strings"

	_ "github.com/aqua777/ai-nexus/dotenv"
)

func main() {
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "TEST_") {
			fmt.Println(env)
		}
	}
}

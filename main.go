package main

import (
	"fmt"
	"os"

	"github.com/Clever/microplane/cmd"
)

// version is set during build, see Makefile's `build` target
var version string

func main() {
	if err := cmd.Execute(version); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

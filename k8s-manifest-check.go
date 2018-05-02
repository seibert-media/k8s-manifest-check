package main

import (
	"flag"
	"os"
	"fmt"
)

func main() {
	flag.Parse()
	if len(flag.Args()) == 0 {
		fmt.Println("missing arg")
		os.Exit(1)
	}
}

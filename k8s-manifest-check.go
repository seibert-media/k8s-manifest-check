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
	for _, arg := range flag.Args() {
		if _, err := os.Stat(arg); os.IsNotExist(err) {
			fmt.Printf("manifest %s not found\n", arg)
			os.Exit(1)
		}
	}
}

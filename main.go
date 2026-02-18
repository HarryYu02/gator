package main

import (
	"fmt"
	"os"

	"github.com/harryyu02/gator/internal/config"
)

func printErrorAndExit(err error) {
	fmt.Printf("err: %v\n", err)
	os.Exit(1)
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		printErrorAndExit(err)
	}

	cfg.SetUser("Harry")

	cfg, err = config.Read()
	if err != nil {
		printErrorAndExit(err)
	}
}

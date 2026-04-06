// Package main is the entry point for the bt CLI.
package main

import (
	"os"

	"github.com/tesserabox/bentotask/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}

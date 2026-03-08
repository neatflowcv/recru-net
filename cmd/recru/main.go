package main

import (
	"fmt"
	"os"

	"github.com/neatflowcv/recru-net/internal/cli"
)

func main() {
	err := cli.Run(os.Args[1:], os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

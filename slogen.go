package main

import (
	"fmt"
	"io"
	"os"

	"github.com/ajalab/slogen/cmd"
)

func run(args []string, stdout, stderr io.Writer) error {
	return cmd.Execute(args, stdout, stderr)
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

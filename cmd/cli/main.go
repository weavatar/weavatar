package main

import (
	"fmt"
	"os"

	_ "time/tzdata"
)

func main() {
	cli, err := initCli()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err = cli.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

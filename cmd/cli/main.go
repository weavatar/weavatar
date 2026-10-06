package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	_ "time/tzdata"

	"github.com/weavatar/weavatar/internal/app"
)

// version is set at build time with -ldflags "-X main.version=v1.2.3".
var version = "dev"

// Errors go to stderr: the app logger's writer is already closed here.
func main() {
	// pgx scans timestamptz into time.Local; responses and logs stay in UTC
	time.Local = time.UTC

	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run() (err error) {
	cli, cleanup, err := app.InitializeCLI()
	if err != nil {
		return err
	}
	defer func() {
		if cleanup != nil {
			err = errors.Join(err, cleanup())
		}
	}()

	return cli.Run(version)
}

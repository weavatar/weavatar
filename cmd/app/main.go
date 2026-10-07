package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "time/tzdata"

	"github.com/libtnb/graceful"
	"github.com/weavatar/weavatar/internal/app"
)

// version is injected at build time: -ldflags "-X main.version=v1.2.3".
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

// run fails only when the service itself failed. Trouble while stopping is
// reported but exits zero: after an upgrade handoff this process's exit
// status becomes the unit's result, and a failure there makes systemd restart
// the service on the next reload.
func run() error {
	fmt.Println("[APP] version", version)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	application, cleanup, err := app.InitializeApp(version)
	if err != nil {
		return err
	}
	err = application.Run(ctx)
	if cleanupErr := cleanup(); cleanupErr != nil {
		fmt.Fprintln(os.Stderr, "Warning:", cleanupErr)
	}
	// a bare assertion on purpose: a drain failure joined onto a real cause stays a failure
	if drain, ok := err.(*graceful.DrainError); ok { //nolint:errorlint
		fmt.Fprintln(os.Stderr, "Warning:", drain)
		return nil
	}

	return err
}

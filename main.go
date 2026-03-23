package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/osir/cli/cmd"
)

var version = "1.0.1"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cmd.SetVersion(version)
	if err := cmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ButterHost69/sloper/internal/runtime"
	"github.com/ButterHost69/sloper/internal/version"
)

func start() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			fmt.Println("sloper", version.Version)
			os.Exit(0)
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: cannot get working directory:", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	rt := runtime.NewRuntime(cwd)
	rt.Start(ctx)
}

func main() {
	start()
}

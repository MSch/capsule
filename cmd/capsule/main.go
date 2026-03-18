package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/MSch/capsule/internal/setup"
)

func main() {
	service := setup.NewService(
		setup.NewConsolePrompter(os.Stdin, os.Stdout),
		setup.NewExecRunner(os.Stdout, os.Stderr),
		setup.NewHostDetector(),
		os.Stdout,
	)

	if err := run(context.Background(), os.Args[1:], os.Stdout, service); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, out io.Writer, service *setup.Service) error {
	if len(args) == 0 {
		printUsage(out)
		return fmt.Errorf("missing command")
	}

	switch args[0] {
	case "help", "-h", "--help":
		printUsage(out)
		return nil
	case "setup":
		if len(args) != 1 {
			printUsage(out)
			return fmt.Errorf("unexpected arguments for setup")
		}

		return service.Run(ctx)
	default:
		printUsage(out)
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func printUsage(out io.Writer) {
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  capsule setup")
}

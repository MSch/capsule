package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/MSch/capsule/internal/setup"
)

func main() {
	if err := run(context.Background(), os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, in io.Reader, out, errOut io.Writer) error {
	if len(args) == 0 {
		printUsage(out)
		return fmt.Errorf("missing command")
	}

	service := setup.NewService(
		setup.NewConsolePrompter(in, out),
		setup.NewExecRunner(out, errOut),
		setup.NewHostDetector(),
		out,
	)

	switch args[0] {
	case "help", "-h", "--help":
		printUsage(out)
		return nil
	case "__bootstrap-local-linux-server":
		if len(args) != 1 {
			return fmt.Errorf("unexpected arguments for __bootstrap-local-linux-server")
		}

		return setup.BootstrapLocalLinuxServer(ctx)
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

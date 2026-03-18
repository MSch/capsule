package app

import (
	"context"
	"fmt"
	"io"

	"github.com/MSch/capsule/internal/setup"
)

type App struct {
	service *setup.Service
	out     io.Writer
}

func New(in io.Reader, out, errOut io.Writer) *App {
	return &App{
		service: setup.NewService(
			setup.NewConsolePrompter(in, out),
			setup.NewExecRunner(out, errOut),
			setup.NewHostDetector(),
			out,
		),
		out: out,
	}
}

func (a *App) Run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		a.printUsage()
		return fmt.Errorf("missing command")
	}

	switch args[0] {
	case "help", "-h", "--help":
		a.printUsage()
		return nil
	case "setup":
		if len(args) != 1 {
			a.printUsage()
			return fmt.Errorf("unexpected arguments for setup")
		}

		return a.service.Run(ctx)
	default:
		a.printUsage()
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func (a *App) printUsage() {
	fmt.Fprintln(a.out, "Usage:")
	fmt.Fprintln(a.out, "  capsule setup")
}

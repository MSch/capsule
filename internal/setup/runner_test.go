package setup

import (
	"context"
	"io"
	"testing"
)

func TestExecRunnerRunSetsEnv(t *testing.T) {
	t.Parallel()

	runner := NewExecRunner(io.Discard, io.Discard)
	result, err := runner.Run(context.Background(), CommandSpec{
		Name: "sh",
		Args: []string{"-c", `printf %s "$COLIMA_PROFILE"`},
		Env:  []string{"COLIMA_PROFILE=capsule"},
	})
	if err != nil {
		t.Fatalf("Run returned an error: %v", err)
	}
	if result.Stdout != "capsule" {
		t.Fatalf("expected COLIMA_PROFILE to be propagated, got %q", result.Stdout)
	}
}

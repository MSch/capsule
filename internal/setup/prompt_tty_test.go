package setup

import (
	"errors"
	"io"
	"testing"

	"github.com/chzyer/readline"
)

func TestNormalizeReadlineError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		err  error
		want error
	}{
		{name: "nil", err: nil, want: nil},
		{name: "interrupt", err: readline.ErrInterrupt, want: errPromptCancelled},
		{name: "eof", err: io.EOF, want: errPromptCancelled},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := normalizeReadlineError(testCase.err)
			if !errors.Is(got, testCase.want) {
				t.Fatalf("expected %v, got %v", testCase.want, got)
			}
		})
	}
}

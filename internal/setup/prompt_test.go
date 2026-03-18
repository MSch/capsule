package setup

import (
	"strings"
	"testing"
)

func TestConsolePrompterAskUsesDedicatedPromptLine(t *testing.T) {
	t.Parallel()

	var out strings.Builder
	prompt := NewConsolePrompter(strings.NewReader("root@example.com\n"), &out)

	answer, err := prompt.Ask("SSH target for the remote server (for example: root@203.0.113.10)", "")
	if err != nil {
		t.Fatalf("Ask returned an error: %v", err)
	}

	if answer != "root@example.com" {
		t.Fatalf("unexpected answer %q", answer)
	}

	expected := "SSH target for the remote server (for example: root@203.0.113.10)\n→ "
	if out.String() != expected {
		t.Fatalf("unexpected prompt output %q", out.String())
	}
}

func TestConsolePrompterSelectSeparatesPrompts(t *testing.T) {
	t.Parallel()

	var out strings.Builder
	prompt := NewConsolePrompter(strings.NewReader("2\n"), &out)

	choice, err := prompt.Select("Where do you want to set up Incus?", []string{
		"Install locally",
		"Connect to a remote Debian/Ubuntu host over SSH",
	})
	if err != nil {
		t.Fatalf("Select returned an error: %v", err)
	}

	if choice != 1 {
		t.Fatalf("unexpected choice %d", choice)
	}

	expected := strings.Join([]string{
		"Where do you want to set up Incus?",
		"  1. Install locally",
		"  2. Connect to a remote Debian/Ubuntu host over SSH",
		"Choose an option [1-2]",
		"→ ",
	}, "\n")

	if out.String() != expected {
		t.Fatalf("unexpected prompt output %q", out.String())
	}
}

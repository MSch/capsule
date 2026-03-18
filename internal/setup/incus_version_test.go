package setup

import "testing"

func TestParseIncusVersionMatched(t *testing.T) {
	t.Parallel()

	version := ParseIncusVersion("Client version: 6.10\nServer version: 6.10\n")
	if version.Client != "6.10" {
		t.Fatalf("expected client version, got %q", version.Client)
	}
	if version.Server != "6.10" {
		t.Fatalf("expected server version, got %q", version.Server)
	}
	if !version.HasServer {
		t.Fatal("expected server to be detected")
	}
	if !version.Matches {
		t.Fatal("expected versions to match")
	}
}

func TestParseIncusVersionWithoutServer(t *testing.T) {
	t.Parallel()

	version := ParseIncusVersion("Client version: 6.10\nError: cannot connect to the server\n")
	if version.HasServer {
		t.Fatal("expected no server to be detected")
	}
	if version.Matches {
		t.Fatal("expected versions not to match")
	}
}

func TestLastNonEmptyLine(t *testing.T) {
	t.Parallel()

	line := lastNonEmptyLine("header line\n\ntoken-value\n")
	if line != "token-value" {
		t.Fatalf("expected last non-empty line to be token-value, got %q", line)
	}
}

package setup

import (
	"bufio"
	"strings"
)

type IncusVersion struct {
	Client    string
	Server    string
	HasServer bool
	Matches   bool
}

func ParseIncusVersion(output string) IncusVersion {
	version := IncusVersion{}
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lower := strings.ToLower(line)

		switch {
		case strings.HasPrefix(lower, "client version:"):
			version.Client = strings.TrimSpace(strings.TrimPrefix(line, "Client version:"))
		case strings.HasPrefix(lower, "server version:"):
			version.Server = strings.TrimSpace(strings.TrimPrefix(line, "Server version:"))
		}
	}

	version.HasServer = version.Server != ""
	version.Matches = version.HasServer && version.Client != "" && version.Client == version.Server

	return version
}

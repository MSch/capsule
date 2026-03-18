package setup

import (
	_ "embed"
	"fmt"
	"os"
)

//go:embed scripts/incus-linux.sh
var linuxInstallerScript string

func writeInstallerScript() (string, func(), error) {
	file, err := os.CreateTemp("", "capsule-incus-*.sh")
	if err != nil {
		return "", nil, fmt.Errorf("creating installer script: %w", err)
	}

	if _, err := file.WriteString(linuxInstallerScript); err != nil {
		file.Close()
		os.Remove(file.Name())
		return "", nil, fmt.Errorf("writing installer script: %w", err)
	}

	if err := file.Close(); err != nil {
		os.Remove(file.Name())
		return "", nil, fmt.Errorf("closing installer script: %w", err)
	}

	if err := os.Chmod(file.Name(), 0o755); err != nil {
		os.Remove(file.Name())
		return "", nil, fmt.Errorf("marking installer script executable: %w", err)
	}

	cleanup := func() {
		_ = os.Remove(file.Name())
	}

	return file.Name(), cleanup, nil
}

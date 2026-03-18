package setup

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"runtime"
	"strings"
)

type Host struct {
	GOOS     string
	Distro   string
	Hostname string
	Username string
}

func (h Host) IsDebianLike() bool {
	return h.Distro == "debian" || h.Distro == "ubuntu"
}

type HostDetector interface {
	Detect() (Host, error)
}

type hostDetector struct{}

func NewHostDetector() HostDetector {
	return hostDetector{}
}

func (hostDetector) Detect() (Host, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return Host{}, fmt.Errorf("detecting hostname: %w", err)
	}

	currentUser, err := user.Current()
	if err != nil {
		return Host{}, fmt.Errorf("detecting current user: %w", err)
	}

	host := Host{
		GOOS:     runtime.GOOS,
		Hostname: hostname,
		Username: currentUser.Username,
	}

	if runtime.GOOS != "linux" {
		return host, nil
	}

	distro, err := readLinuxDistro("/etc/os-release")
	if err != nil {
		return Host{}, err
	}
	host.Distro = distro

	return host, nil
}

func readLinuxDistro(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "ID=") {
			continue
		}

		return strings.Trim(strings.TrimPrefix(line, "ID="), `"`), nil
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scanning %s: %w", path, err)
	}

	return "", fmt.Errorf("could not detect Linux distribution from %s", path)
}

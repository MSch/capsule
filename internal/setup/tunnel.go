package setup

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type SocketTunnel interface {
	Path() string
	Close() error
}

type SocketTunnelOpener interface {
	Open(ctx context.Context, target, remoteSocket string, sshOptions []string) (SocketTunnel, error)
}

type sshSocketTunnelOpener struct{}

func newSocketTunnelOpener() SocketTunnelOpener {
	return sshSocketTunnelOpener{}
}

func (sshSocketTunnelOpener) Open(ctx context.Context, target, remoteSocket string, sshOptions []string) (SocketTunnel, error) {
	socketFile, err := os.CreateTemp("", "capsule-incus-*.sock")
	if err != nil {
		return nil, fmt.Errorf("creating a local tunnel socket path: %w", err)
	}

	localSocket := socketFile.Name()
	if err := socketFile.Close(); err != nil {
		_ = os.Remove(localSocket)
		return nil, fmt.Errorf("closing the local tunnel socket placeholder: %w", err)
	}

	if err := os.Remove(localSocket); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("removing the local tunnel socket placeholder: %w", err)
	}

	args := append([]string{}, sshOptions...)
	args = append(args,
		"-o", "ExitOnForwardFailure=yes",
		"-o", "StreamLocalBindUnlink=yes",
		"-N",
		"-L", fmt.Sprintf("%s:%s", localSocket, remoteSocket),
		target,
	)

	cmd := exec.CommandContext(ctx, "ssh", args...)

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Start(); err != nil {
		_ = os.Remove(localSocket)
		return nil, fmt.Errorf("starting the SSH tunnel: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, err := os.Stat(localSocket); err == nil {
			return &sshSocketTunnel{
				localSocket: localSocket,
				cmd:         cmd,
				done:        done,
			}, nil
		}

		select {
		case err := <-done:
			_ = os.Remove(localSocket)
			message := strings.TrimSpace(output.String())
			if message != "" {
				return nil, fmt.Errorf("starting the SSH tunnel: %w\n%s", err, message)
			}

			return nil, fmt.Errorf("starting the SSH tunnel: %w", err)
		default:
		}

		if time.Now().After(deadline) {
			_ = cmd.Process.Kill()
			<-done
			_ = os.Remove(localSocket)
			return nil, errors.New("timed out while waiting for the SSH tunnel to open")
		}

		timer := time.NewTimer(100 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			_ = cmd.Process.Kill()
			<-done
			_ = os.Remove(localSocket)
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}

type sshSocketTunnel struct {
	localSocket string
	cmd         *exec.Cmd
	done        <-chan error
}

func (t *sshSocketTunnel) Path() string {
	return t.localSocket
}

func (t *sshSocketTunnel) Close() error {
	if t.cmd.Process != nil {
		_ = t.cmd.Process.Kill()
	}

	err := <-t.done
	_ = os.Remove(t.localSocket)

	var exitErr *exec.ExitError
	if err == nil || errors.As(err, &exitErr) {
		return nil
	}

	return err
}

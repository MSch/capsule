# capsule

Single-module Go monorepo for the `capsule` CLI.

## Layout

- `cmd/cli`: main package for the `capsule` binary
- `internal/setup`: interactive setup flows for local and remote Incus installs

## Commands

```bash
go run ./cmd/cli --help
go run ./cmd/cli setup
```

The `capsule setup` flow currently supports:

- local setup on macOS using `colima` + `incus`
- local setup on Debian/Ubuntu using the Zabbly Incus packages
- remote Debian/Ubuntu server setup over SSH, including local client wiring

Remote automation expects SSH access as `root` or as a user with passwordless `sudo`.

## Developer Workflow

```bash
make bootstrap
make fmt
make lint
make test
make check
```

`make bootstrap` installs the pinned formatter and linter into `.bin/` and configures Git to use the repository's `.githooks` directory. The pre-commit hook runs:

- `make fmt-check`
- `make lint`
- `make test`

## Releases

Tag pushes that match `v*` trigger `.github/workflows/release.yml`, which publishes tarballs for:

- `linux/amd64`
- `linux/arm64`
- `darwin/amd64`
- `darwin/arm64`

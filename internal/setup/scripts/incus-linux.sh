#!/usr/bin/env bash
set -euo pipefail

MODE="server"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --mode=*)
      MODE="${1#*=}"
      shift
      ;;
    *)
      echo "unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

if [[ "$(id -u)" -ne 0 ]]; then
  echo "this installer must run as root" >&2
  exit 1
fi

if [[ ! -f /etc/os-release ]]; then
  echo "expected /etc/os-release on this Linux host" >&2
  exit 1
fi

# shellcheck disable=SC1091
source /etc/os-release

if [[ "${ID:-}" != "debian" && "${ID:-}" != "ubuntu" ]]; then
  echo "this installer currently supports Debian and Ubuntu only" >&2
  exit 1
fi

ARCH="$(dpkg --print-architecture)"
CODENAME="${VERSION_CODENAME:-}"
if [[ -z "$CODENAME" ]]; then
  echo "could not determine the distro codename from /etc/os-release" >&2
  exit 1
fi

export DEBIAN_FRONTEND=noninteractive

apt-get update
apt-get install -y ca-certificates curl gpg

install -d -m 0755 /etc/apt/keyrings
curl -fsSL https://pkgs.zabbly.com/key.asc -o /etc/apt/keyrings/zabbly.asc

cat >/etc/apt/sources.list.d/zabbly-incus-stable.sources <<EOF
Enabled: yes
Types: deb
URIs: https://pkgs.zabbly.com/incus/stable
Suites: ${CODENAME}
Components: main
Architectures: ${ARCH}
Signed-By: /etc/apt/keyrings/zabbly.asc
EOF

apt-get update

PACKAGE="incus"
if [[ "$MODE" == "client" ]]; then
  PACKAGE="incus-client"
fi

apt-get install -y "${PACKAGE}"

if [[ "$MODE" != "server" ]]; then
  exit 0
fi

if command -v systemctl >/dev/null 2>&1; then
  systemctl enable --now incus.service >/dev/null 2>&1 || systemctl restart incus.service
fi

incus admin waitready >/dev/null 2>&1 || true

if ! incus storage list --format csv -c n >/dev/null 2>&1 || [[ -z "$(incus storage list --format csv -c n 2>/dev/null)" ]]; then
  incus admin init --auto --storage-backend dir
fi

incus config set core.https_address :8443

if command -v ufw >/dev/null 2>&1 && ufw status | grep -q "Status: active"; then
  ufw allow 8443/tcp >/dev/null 2>&1 || true
fi

TARGET_USER="${SUDO_USER:-${USER:-root}}"
if [[ -n "$TARGET_USER" && "$TARGET_USER" != "root" ]] && getent group incus-admin >/dev/null 2>&1; then
  usermod -aG incus-admin "$TARGET_USER" || true
fi

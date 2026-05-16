#!/usr/bin/env bash
# Downloads the geoagenda CLI binary that matches the host OS/arch and
# installs it under the plugin's bin/ directory.
#
# Triggered by the SessionStart hook on every Claude Code session. Fast path:
# if the binary is already present, exit 0 immediately. To force a refresh,
# delete the binary (or set GEOAGENDA_CLI_FORCE_INSTALL=1) and start a new
# session.
set -euo pipefail

REPO="scalegems/geoagenda-cli"
PROJECT_NAME="geoagenda"

PLUGIN_ROOT="${CLAUDE_PLUGIN_ROOT:-$(cd "$(dirname "$0")/.." && pwd)}"
BIN_DIR="${PLUGIN_ROOT}/bin"
mkdir -p "${BIN_DIR}"

# Fast path — skip if the binary is already installed.
case "$(uname -s)" in
  MINGW*|MSYS*|CYGWIN*) EXISTING_BIN="${BIN_DIR}/${PROJECT_NAME}.exe" ;;
  *) EXISTING_BIN="${BIN_DIR}/${PROJECT_NAME}" ;;
esac
if [[ -x "$EXISTING_BIN" && -z "${GEOAGENDA_CLI_FORCE_INSTALL:-}" ]]; then
  exit 0
fi

# Normalize OS
case "$(uname -s)" in
  Darwin)  OS=darwin  ;;
  Linux)   OS=linux   ;;
  MINGW*|MSYS*|CYGWIN*) OS=windows ;;
  *) echo "unsupported OS: $(uname -s)" >&2; exit 1 ;;
esac

# Normalize arch to GoReleaser conventions
case "$(uname -m)" in
  x86_64|amd64)        ARCH=amd64 ;;
  arm64|aarch64)       ARCH=arm64 ;;
  *) echo "unsupported arch: $(uname -m)" >&2; exit 1 ;;
esac

# Windows arm64 is excluded from the release matrix
if [[ "$OS" == "windows" && "$ARCH" == "arm64" ]]; then
  echo "windows/arm64 is not built; install from source via 'go build'" >&2
  exit 1
fi

EXT="tar.gz"
[[ "$OS" == "windows" ]] && EXT="zip"

ASSET="${PROJECT_NAME}_${OS}_${ARCH}.${EXT}"
TAG=$(curl -fsSLI -o /dev/null -w '%{url_effective}' \
  "https://github.com/${REPO}/releases/latest" \
  | sed 's|.*/tag/||' \
  | tr -d '\r\n')

if [[ -z "$TAG" ]]; then
  echo "could not resolve latest release tag from github.com/${REPO}" >&2
  exit 1
fi

URL="https://github.com/${REPO}/releases/download/${TAG}/${ASSET}"
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

echo "Fetching ${ASSET} (${TAG})..."
if ! curl -fsSL "$URL" -o "${TMP}/${ASSET}"; then
  echo "failed to download ${URL}" >&2
  exit 1
fi

if [[ "$EXT" == "zip" ]]; then
  unzip -q "${TMP}/${ASSET}" -d "${TMP}"
else
  tar -xzf "${TMP}/${ASSET}" -C "${TMP}"
fi

BIN_NAME="${PROJECT_NAME}"
[[ "$OS" == "windows" ]] && BIN_NAME="${PROJECT_NAME}.exe"

if [[ ! -f "${TMP}/${BIN_NAME}" ]]; then
  echo "archive did not contain ${BIN_NAME}" >&2
  ls -la "${TMP}" >&2
  exit 1
fi

install -m 0755 "${TMP}/${BIN_NAME}" "${BIN_DIR}/${BIN_NAME}"
echo "Installed ${BIN_DIR}/${BIN_NAME} (${TAG})"

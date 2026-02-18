#!/usr/bin/env sh
set -eu

REPO="alexisprovost/immprune"
INSTALL_DIR="${IMMPRUNE_INSTALL_DIR:-/usr/local/bin}"
REQUESTED_VERSION="${1:-latest}"

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "Error: '$1' is required" >&2
    exit 1
  }
}

need_cmd curl
need_cmd tar

os_name() {
  case "$(uname -s)" in
    Darwin) echo "darwin" ;;
    *)
      echo "Error: unsupported OS '$(uname -s)'. This installer supports macOS only." >&2
      exit 1
      ;;
  esac
}

arch_name() {
  case "$(uname -m)" in
    aarch64|arm64) echo "arm64" ;;
    *)
      echo "Error: unsupported architecture '$(uname -m)'. This installer supports Apple Silicon (arm64) only." >&2
      exit 1
      ;;
  esac
}

OS="$(os_name)"
ARCH="$(arch_name)"
ASSET_BASENAME="immprune-${OS}-${ARCH}"
ARCHIVE_NAME="${ASSET_BASENAME}.tar.gz"

if [ "$REQUESTED_VERSION" = "latest" ]; then
  DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/${ARCHIVE_NAME}"
else
  case "$REQUESTED_VERSION" in
    v*) TAG="$REQUESTED_VERSION" ;;
    *) TAG="v$REQUESTED_VERSION" ;;
  esac
  DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${TAG}/${ARCHIVE_NAME}"
fi

TMP_DIR="$(mktemp -d)"
cleanup() {
  rm -rf "$TMP_DIR"
}
trap cleanup EXIT INT TERM

ARCHIVE_PATH="$TMP_DIR/$ARCHIVE_NAME"
echo "Downloading $DOWNLOAD_URL"
curl -fsSL "$DOWNLOAD_URL" -o "$ARCHIVE_PATH"

tar -xzf "$ARCHIVE_PATH" -C "$TMP_DIR"
BINARY_PATH="$TMP_DIR/$ASSET_BASENAME"

if [ ! -f "$BINARY_PATH" ]; then
  echo "Error: downloaded archive does not contain expected binary '$ASSET_BASENAME'." >&2
  exit 1
fi

if [ -w "$INSTALL_DIR" ]; then
  SUDO=""
else
  need_cmd sudo
  SUDO="sudo"
fi

$SUDO mkdir -p "$INSTALL_DIR"
$SUDO install -m 0755 "$BINARY_PATH" "$INSTALL_DIR/immprune"

echo "immprune installed at $INSTALL_DIR/immprune"
"$INSTALL_DIR/immprune" --help >/dev/null 2>&1 || true

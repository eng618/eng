#!/bin/sh
# eng CLI installer — POSIX sh, only requires curl + tar (unzip on Windows).
# Installs the latest (or pinned) release from GitHub Releases into INSTALL_DIR.
#
# Usage:
#   curl -sSfL https://raw.githubusercontent.com/eng618/eng/main/install.sh | sh
#   curl -sSfL .../install.sh | sh -s -- -v v0.17.5 --to "$HOME/.local/bin"
#
# Env overrides:
#   ENG_VERSION  version tag to install (e.g. v0.17.5) or "latest" (default)
#   INSTALL_DIR  destination directory (default: /usr/local/bin)
#   REPO         GitHub repo "owner/name" (default: eng618/eng)
set -eu

REPO="${REPO:-eng618/eng}"
VERSION_INPUT="${ENG_VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
USE_SUDO="auto"

log() { printf '%s\n' "eng-install: $*" >&2; }
err() { printf '%s\n' "eng-install: ERROR: $*" >&2; }
die() { err "$*"; exit 1; }

usage() {
  cat >&2 <<'EOF'
Usage: install.sh [-v VERSION] [-b DIR] [--to DIR] [-h]

  -v, --version VERSION   Release tag to install (e.g. v0.17.5) or "latest" (default).
                          Can also be set via ENG_VERSION env var.
  -b, --to, --install-dir DIR
                          Destination directory (default: /usr/local/bin).
                          Can also be set via INSTALL_DIR env var.
  -h, --help              Show this help and exit.

Examples:
  curl -sSfL https://raw.githubusercontent.com/eng618/eng/main/install.sh | sh
  curl -sSfL .../install.sh | sh -s -- -v v0.17.5
  INSTALL_DIR="$HOME/.local/bin" sh install.sh
EOF
}

# --- arg parsing (POSIX, supports -v X, -v=X, --version X, --version=X, -b/--to/--install-dir) ---
while [ $# -gt 0 ]; do
  case "$1" in
    -h|--help) usage; exit 0 ;;
    -v|--version)
      [ $# -ge 2 ] || die "Missing value for $1 (e.g. $1 v0.17.5)"
      VERSION_INPUT="$2"; shift 2 ;;
    -v=*|--version=*)
      VERSION_INPUT="${1#*=}"; shift ;;
    -b|--to|--install-dir)
      [ $# -ge 2 ] || die "Missing value for $1 (e.g. $1 \$HOME/.local/bin)"
      INSTALL_DIR="$2"; shift 2 ;;
    -b=*|--to=*|--install-dir=*)
      INSTALL_DIR="${1#*=}"; shift ;;
    --) shift; break ;;
    -*) die "Unknown option: $1 (see --help)" ;;
    *) break ;;
  esac
done

command -v curl >/dev/null 2>&1 || die "curl is required but not installed"
command -v tar >/dev/null 2>&1 || die "tar is required but not installed"

# --- platform detection (must match .goreleaser.yaml name_template) ---
OS="$(uname -s)"
case "$OS" in
  Darwin) OS_LABEL="Darwin" ;;
  Linux) OS_LABEL="Linux" ;;
  MINGW*|MSYS*|CYGWIN*|Windows*) OS_LABEL="Windows" ;;
  *) die "Unsupported OS: $OS (only macOS, Linux, and Windows Git Bash are supported)" ;;
esac

ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64) ARCH_LABEL="x86_64" ;;
  arm64|aarch64) ARCH_LABEL="arm64" ;;
  i386|i686) ARCH_LABEL="i386" ;;
  *) die "Unsupported architecture: $ARCH (only x86_64, arm64, i386 are supported)" ;;
esac

# --- version resolution ---
resolve_latest_tag() {
  # Prefer redirect-based lookup (no API rate limits, no jq).
  final_url="$(curl -sSfL -o /dev/null -w '%{url_effective}' "https://github.com/${REPO}/releases/latest" 2>/dev/null || true)"
  case "$final_url" in
    */tag/v*|*/tag/[0-9]*)
      tag="${final_url##*/tag/}"
      [ -n "$tag" ] && { printf '%s' "$tag"; return 0; }
      ;;
  esac
  # Fallback: GitHub API (grep/sed only, no jq).
  api_json="$(curl -sSfL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null || true)"
  tag="$(printf '%s' "$api_json" | grep -m1 '"tag_name"' | sed -e 's/.*"tag_name"[[:space:]]*:[[:space:]]*"//' -e 's/".*//' || true)"
  [ -n "$tag" ] || return 1
  printf '%s' "$tag"
}

case "$VERSION_INPUT" in
  latest|""|LATEST)
    log "Resolving latest release for ${REPO}..."
    TAG="$(resolve_latest_tag)" || die "Could not resolve latest release (check network or set ENG_VERSION=vX.Y.Z)"
    ;;
  v*) TAG="$VERSION_INPUT" ;;
  [0-9]*) TAG="v$VERSION_INPUT" ;;
  *) die "Invalid version: $VERSION_INPUT (expected vX.Y.Z or 'latest')" ;;
esac
# GoReleaser strips the leading 'v' for the asset filename, keeps it for the tag path.
VERSION_NO_V="${TAG#v}"
log "Installing eng ${TAG} for ${OS_LABEL}/${ARCH_LABEL}..."

if [ "$OS_LABEL" = "Windows" ]; then
  ASSET="eng_${VERSION_NO_V}_Windows_${ARCH_LABEL}.zip"
else
  ASSET="eng_${VERSION_NO_V}_${OS_LABEL}_${ARCH_LABEL}.tar.gz"
fi
BASE_URL="https://github.com/${REPO}/releases/download/${TAG}"

TMPDIR="$(mktemp -d 2>/dev/null || mktemp -d -t eng-install)" || die "Could not create temp directory"
trap 'rm -rf "$TMPDIR"' EXIT INT TERM

log "Downloading ${ASSET}..."
curl -sSfL --retry 3 -o "$TMPDIR/$ASSET" "$BASE_URL/$ASSET" \
  || die "Download failed: $BASE_URL/$ASSET (check that version/platform exists)"

log "Downloading checksums.txt..."
curl -sSfL --retry 3 -o "$TMPDIR/checksums.txt" "$BASE_URL/checksums.txt" \
  || die "Download failed: $BASE_URL/checksums.txt"

# --- checksum verification (sha256sum preferred, shasum fallback) ---
EXPECTED="$(grep -F "  $ASSET" "$TMPDIR/checksums.txt" | awk '{print $1}' || true)"
[ -n "$EXPECTED" ] || die "Checksum entry for $ASSET not found in checksums.txt"
if command -v sha256sum >/dev/null 2>&1; then
  ACTUAL="$(sha256sum "$TMPDIR/$ASSET" | awk '{print $1}')"
elif command -v shasum >/dev/null 2>&1; then
  ACTUAL="$(shasum -a 256 "$TMPDIR/$ASSET" | awk '{print $1}')"
else
  die "Neither sha256sum nor shasum found — cannot verify checksum"
fi
[ "$ACTUAL" = "$EXPECTED" ] || die "Checksum mismatch for $ASSET (expected $EXPECTED, got $ACTUAL)"
log "Checksum verified."

# --- extraction ---
if [ "$OS_LABEL" = "Windows" ]; then
  command -v unzip >/dev/null 2>&1 || die "unzip is required on Windows but not installed"
  (cd "$TMPDIR" && unzip -oq "$ASSET") || die "Failed to extract $ASSET"
  SRC_BIN="$TMPDIR/eng.exe"
  [ -f "$SRC_BIN" ] || SRC_BIN="$TMPDIR/eng"
  DEST_NAME="eng.exe"
else
  tar -xzf "$TMPDIR/$ASSET" -C "$TMPDIR" || die "Failed to extract $ASSET"
  SRC_BIN="$TMPDIR/eng"
  [ -f "$SRC_BIN" ] || die "Extracted archive did not contain an 'eng' binary"
  DEST_NAME="eng"
fi

# --- install (sudo only when needed) ---
mkdir -p "$INSTALL_DIR" 2>/dev/null || {
  if command -v sudo >/dev/null 2>&1; then
    sudo mkdir -p "$INSTALL_DIR" || die "Could not create $INSTALL_DIR"
  else
    die "Could not create $INSTALL_DIR (permission denied and sudo not available)"
  fi
}

DEST="$INSTALL_DIR/$DEST_NAME"
if [ -w "$INSTALL_DIR" ]; then
  USE_SUDO=""
elif command -v sudo >/dev/null 2>&1 && [ "$USE_SUDO" = "auto" ]; then
  USE_SUDO="sudo"
else
  die "No write permission for $INSTALL_DIR (re-run with sudo or set INSTALL_DIR=\$HOME/.local/bin)"
fi

if [ -n "$USE_SUDO" ]; then
  if ! $USE_SUDO install -m 755 "$SRC_BIN" "$DEST" 2>/dev/null; then
    $USE_SUDO cp -f "$SRC_BIN" "$DEST" || die "Install to $DEST failed"
    $USE_SUDO chmod 755 "$DEST" || die "Install to $DEST failed"
  fi
else
  if command -v install >/dev/null 2>&1; then
    if ! install -m 755 "$SRC_BIN" "$DEST" 2>/dev/null; then
      cp -f "$SRC_BIN" "$DEST" || die "Install to $DEST failed"
      chmod 755 "$DEST" || die "chmod $DEST failed"
    fi
  else
    cp -f "$SRC_BIN" "$DEST" || die "Install to $DEST failed"
    chmod 755 "$DEST" || die "chmod $DEST failed"
  fi
fi

# --- smoke test ---
if [ -x "$DEST" ]; then
  log "Installed to $DEST"
  if "$DEST" --version >/dev/null 2>&1; then
    log "Verified: $("$DEST" --version 2>/dev/null | head -n 1)"
  fi
else
  die "Install reported success but $DEST is not executable"
fi

case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *) log "WARNING: $INSTALL_DIR is not on your PATH. Add: export PATH=\"$INSTALL_DIR:\$PATH\"" ;;
esac

log "Done. Run 'eng --help' to get started. Update anytime with 'eng version -u'."

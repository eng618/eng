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
#   ENG_COMPLETIONS     1/0 to enable/disable shell completions (default: 1)
#   ENG_NO_COMPLETIONS  set to 1 to skip shell completions
#   COMPLETIONS_DIR     override base dir for installed completions (advanced/testing)
set -eu

REPO="${REPO:-eng618/eng}"
VERSION_INPUT="${ENG_VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
USE_SUDO="auto"
COMPLETIONS="${COMPLETIONS:-${ENG_COMPLETIONS:-yes}}"
case "$COMPLETIONS" in
  0|no|No|NO|false|False|FALSE|off|Off|OFF|n|N) COMPLETIONS="no" ;;
  *) COMPLETIONS="yes" ;;
esac
case "${ENG_NO_COMPLETIONS:-}" in
  ""|0|no|No|NO|false|False|FALSE|off|Off|OFF|n|N) ;;
  *) COMPLETIONS="no" ;;
esac
COMPLETIONS_DIR="${COMPLETIONS_DIR:-}"

log() { printf '%s\n' "eng-install: $*" >&2; }
err() { printf '%s\n' "eng-install: ERROR: $*" >&2; }
die() { err "$*"; exit 1; }

usage() {
  cat >&2 <<'EOF'
Usage: install.sh [-v VERSION] [-b DIR] [--to DIR] [--completions|--no-completions] [-h]

  -v, --version VERSION   Release tag to install (e.g. v0.17.5) or "latest" (default).
                          Can also be set via ENG_VERSION env var.
  -b, --to, --install-dir DIR
                          Destination directory (default: /usr/local/bin).
                          Can also be set via INSTALL_DIR env var.
  --completions           Install shell completions for bash/zsh/fish (default).
  --no-completions        Skip shell completion installation.
                          Can also be disabled via ENG_NO_COMPLETIONS=1
                          or ENG_COMPLETIONS=0 env vars.
  --completions-dir DIR   Override base directory for completions
                          (advanced/testing). Can also be set via
                          COMPLETIONS_DIR env var.
  -h, --help              Show this help and exit.

Examples:
  curl -sSfL https://raw.githubusercontent.com/eng618/eng/main/install.sh | sh
  curl -sSfL .../install.sh | sh -s -- -v v0.17.5
  curl -sSfL .../install.sh | sh -s -- --no-completions
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
    --completions) COMPLETIONS="yes"; shift ;;
    --no-completions) COMPLETIONS="no"; shift ;;
    --completions-dir)
      [ $# -ge 2 ] || die "Missing value for $1 (e.g. $1 \$HOME/.local/share/eng-completions)"
      COMPLETIONS_DIR="$2"; shift 2 ;;
    --completions-dir=*)
      COMPLETIONS_DIR="${1#*=}"; shift ;;
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

# --- shell completions (bash/zsh/fish) ---
# Best-effort: generates completions from the installed binary and drops
# static files into the standard per-shell locations. Never fails the install.
install_one_completion() {
  _shell="$1"
  _target="$2"
  _tmp_comp="$TMPDIR/completion.$_shell"
  if ! "$DEST" completion "$_shell" >"$_tmp_comp" 2>/dev/null; then
    log "WARNING: could not generate $_shell completions (skipping)"
    return 0
  fi
  if [ ! -s "$_tmp_comp" ]; then
    log "WARNING: empty $_shell completions generated (skipping)"
    return 0
  fi
  # Sanity check: the first line must look like a completion script.
  # This refuses output polluted by log lines on stdout (as produced by
  # older binaries). Installing such a file would break shell startup.
  _first_line="$(head -n 1 "$_tmp_comp" 2>/dev/null || true)"
  _valid="no"
  case "$_shell:$_first_line" in
    bash:"# bash completion"*) _valid="yes" ;;
    zsh:"#compdef"*) _valid="yes" ;;
    fish:"# fish completion"*) _valid="yes" ;;
  esac
  if [ "$_valid" != "yes" ]; then
    log "WARNING: generated $_shell completions failed validation (skipping)"
    return 0
  fi
  _target_dir="$(dirname "$_target")"
  _comp_sudo=""
  if ! mkdir -p "$_target_dir" 2>/dev/null; then
    if command -v sudo >/dev/null 2>&1; then
      if ! sudo mkdir -p "$_target_dir" 2>/dev/null; then
        log "WARNING: could not create completions dir $_target_dir (skipping $_shell)"
        return 0
      fi
      _comp_sudo="sudo"
    else
      log "WARNING: no write permission for $_target_dir (skipping $_shell completions)"
      return 0
    fi
  fi
  if [ -n "$_comp_sudo" ]; then
    if ! $_comp_sudo cp -f "$_tmp_comp" "$_target" 2>/dev/null ||
      ! $_comp_sudo chmod 644 "$_target" 2>/dev/null; then
      log "WARNING: could not write $_target (skipping)"
      return 0
    fi
  else
    if command -v install >/dev/null 2>&1; then
      if ! install -m 644 "$_tmp_comp" "$_target" 2>/dev/null; then
        if ! cp -f "$_tmp_comp" "$_target" 2>/dev/null; then
          log "WARNING: could not write $_target (skipping)"
          return 0
        fi
        chmod 644 "$_target" 2>/dev/null || true
      fi
    else
      if ! cp -f "$_tmp_comp" "$_target" 2>/dev/null; then
        log "WARNING: could not write $_target (skipping)"
        return 0
      fi
      chmod 644 "$_target" 2>/dev/null || true
    fi
  fi
  log "Installed $_shell completions to $_target"
  return 0
}

install_completions() {
  if [ "$COMPLETIONS" != "yes" ]; then
    log "Skipping shell completions (--no-completions)."
    return 0
  fi
  if [ ! -x "$DEST" ]; then
    log "WARNING: skipping shell completions ($DEST not executable)"
    return 0
  fi

  # Explicit override (advanced/testing): fixed layout under one base dir.
  if [ -n "$COMPLETIONS_DIR" ]; then
    install_one_completion bash "$COMPLETIONS_DIR/bash/eng"
    install_one_completion zsh "$COMPLETIONS_DIR/zsh/_eng"
    install_one_completion fish "$COMPLETIONS_DIR/fish/eng.fish"
    log "Shell completions installed under $COMPLETIONS_DIR. Restart your shell."
    return 0
  fi

  _home="${HOME:-}"
  _xdg_data="${XDG_DATA_HOME:-}"
  if [ -z "$_xdg_data" ] && [ -n "$_home" ]; then
    _xdg_data="$_home/.local/share"
  fi
  _xdg_config="${XDG_CONFIG_HOME:-}"
  if [ -z "$_xdg_config" ] && [ -n "$_home" ]; then
    _xdg_config="$_home/.config"
  fi

  _bash_user=""
  _zsh_user=""
  _fish_user=""
  if [ -n "$_xdg_data" ]; then
    _bash_user="$_xdg_data/bash-completion/completions"
    _zsh_user="$_xdg_data/zsh/site-functions"
  fi
  if [ -n "$_xdg_config" ]; then
    _fish_user="$_xdg_config/fish/completions"
  fi

  _system_install="no"
  case "$INSTALL_DIR" in
    /usr/local/bin|/opt/homebrew/bin|/usr/bin|/opt/local/bin) _system_install="yes" ;;
  esac

  # On Windows (Git Bash) only user dirs are meaningful; system
  # vendor paths do not apply.
  if [ "$OS_LABEL" = "Windows" ]; then
    if [ -z "$_bash_user" ]; then
      log "WARNING: skipping shell completions (HOME is unset)"
      return 0
    fi
    install_one_completion bash "$_bash_user/eng"
    install_one_completion zsh "$_zsh_user/_eng"
    install_one_completion fish "$_fish_user/eng.fish"
    log "Shell completions installed. Restart your shell."
    return 0
  fi

  _bash_dir="$_bash_user"
  _zsh_dir="$_zsh_user"
  _fish_dir="$_fish_user"
  if [ "$_system_install" = "yes" ]; then
    if [ -n "$_bash_user" ] || [ -n "$_xdg_data" ]; then
      : # user dirs available as fallback
    fi
    _sys_bash="/usr/local/share/bash-completion/completions"
    _sys_zsh="/usr/local/share/zsh/site-functions"
    _sys_fish="/usr/local/share/fish/vendor_completions.d"
    if [ -w "$_sys_bash" ] 2>/dev/null || command -v sudo >/dev/null 2>&1; then
      _bash_dir="$_sys_bash"
    fi
    if [ -w "$_sys_zsh" ] 2>/dev/null || command -v sudo >/dev/null 2>&1; then
      _zsh_dir="$_sys_zsh"
    fi
    if [ -w "$_sys_fish" ] 2>/dev/null || command -v sudo >/dev/null 2>&1; then
      _fish_dir="$_sys_fish"
    fi
  fi

  if [ -z "$_bash_dir" ] && [ -z "$_zsh_dir" ] && [ -z "$_fish_dir" ]; then
    log "WARNING: skipping shell completions (HOME is unset and no system dir is usable)"
    return 0
  fi

  if [ -n "$_bash_dir" ]; then
    install_one_completion bash "$_bash_dir/eng"
  fi
  if [ -n "$_zsh_dir" ]; then
    install_one_completion zsh "$_zsh_dir/_eng"
  fi
  if [ -n "$_fish_dir" ]; then
    install_one_completion fish "$_fish_dir/eng.fish"
  fi

  log "Shell completions installed. Restart your shell (new terminal) to load them."
  log "  zsh:  ensure the _eng file's directory is on fpath before 'compinit'"
  log "  bash: requires bash-completion to auto-load the completions dir"
  log "  fish: completions load automatically"
  return 0
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

install_completions

log "Done. Run 'eng --help' to get started. Update anytime with 'eng version -u'."

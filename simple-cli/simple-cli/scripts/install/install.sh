#!/usr/bin/env sh
# install.sh — POSIX install script for simple-cli on Linux and macOS.
# Constitution Principle III: installer MUST register ENV PATH and be idempotent.
#
# Usage:
#   curl -fsSL https://github.com/binpqh/simple-cli/releases/latest/download/install.sh | sh
#   # or with sudo for system-wide install:
#   sudo sh install.sh

set -e

BINARY="simple-cli"
REPO="binpqh/simple-cli"
INSTALL_DIR_USER="${HOME}/.local/bin"
INSTALL_DIR_SYSTEM="/usr/local/bin"

# ──────────────────────────────────────────────
# Helpers
# ──────────────────────────────────────────────

log()  { printf '\033[32m[install]\033[0m %s\n' "$*" >&2; }
warn() { printf '\033[33m[install]\033[0m %s\n' "$*" >&2; }
err()  { printf '\033[31m[install]\033[0m %s\n' "$*" >&2; exit 1; }

has() { command -v "$1" >/dev/null 2>&1; }

detect_os() {
  case "$(uname -s)" in
    Linux*)  echo "linux" ;;
    Darwin*) echo "darwin" ;;
    *)       err "Unsupported OS: $(uname -s)" ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64)  echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *)             err "Unsupported architecture: $(uname -m)" ;;
  esac
}

download_binary() {
  local os="$1" arch="$2" dest="$3"
  local tag
  tag=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
        | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": "\(.*\)".*/\1/')
  local url="https://github.com/${REPO}/releases/download/${tag}/${BINARY}_${os}_${arch}.tar.gz"
  log "Downloading ${BINARY} ${tag} for ${os}/${arch}..."
  curl -fsSL "$url" | tar -xz -C "$(dirname "$dest")" "${BINARY}"
  chmod +x "$dest"
}

# ──────────────────────────────────────────────
# PATH registration (idempotent)
# ──────────────────────────────────────────────

append_path_if_missing() {
  local file="$1" dir="$2"
  [ -f "$file" ] || return 0
  if ! grep -qF "$dir" "$file" 2>/dev/null; then
    printf '\nexport PATH="$PATH:%s"\n' "$dir" >> "$file"
    log "Added ${dir} to ${file}"
  else
    log "${dir} already in ${file} — skipping"
  fi
}

register_path() {
  local dir="$1"
  append_path_if_missing "${HOME}/.bashrc"  "$dir"
  append_path_if_missing "${HOME}/.zshrc"   "$dir"
  append_path_if_missing "${HOME}/.profile" "$dir"
}

# ──────────────────────────────────────────────
# Main
# ──────────────────────────────────────────────

main() {
  local os arch install_dir
  os=$(detect_os)
  arch=$(detect_arch)

  # Choose install directory based on sudo availability.
  if [ "$(id -u)" = "0" ]; then
    install_dir="$INSTALL_DIR_SYSTEM"
  else
    install_dir="$INSTALL_DIR_USER"
    mkdir -p "$install_dir"
  fi

  local dest="${install_dir}/${BINARY}"
  download_binary "$os" "$arch" "$dest"
  log "Installed ${BINARY} → ${dest}"

  # Register PATH when not a system install.
  if [ "$install_dir" = "$INSTALL_DIR_USER" ]; then
    register_path "$install_dir"
  fi

  # Post-install validation.
  log "Validating installation..."
  if PATH="${install_dir}:${PATH}" "${dest}" --version >/dev/null 2>&1; then
    log "✓ Installation successful: $(PATH="${install_dir}:${PATH}" "${dest}" --version)"
  else
    warn "Binary installed but --version check failed."
    warn "Please open a new terminal and run: ${BINARY} --version"
    warn "If still not found, add ${install_dir} to your PATH manually."
  fi
}

main "$@"

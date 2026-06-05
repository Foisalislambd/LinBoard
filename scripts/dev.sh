#!/usr/bin/env bash
# LinBoard — development helper (build, clean, run, …)
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BINARY="${BINARY:-linboard}"
MAIN_PKG="./cmd/linboard"
LDFLAGS="${LDFLAGS:- -s -w}"

export GOTOOLCHAIN="${GOTOOLCHAIN:-local}"
export GOROOT="${GOROOT:-/usr/local/go}"
export GOPATH="${GOPATH:-$HOME/go}"
export PATH="$GOROOT/bin:$GOPATH/bin:$PATH"

cd "$ROOT"

usage() {
  cat <<'EOF'
LinBoard dev script

Usage:
  ./scripts/dev.sh <command>

Commands:
  build       Build ./linboard binary
  run         Build (if needed) and run LinBoard
  rebuild     Clean binary + full rebuild
  clean       Remove ./linboard binary
  clean-cache Clear Go build cache (slow next build)
  clean-all   Remove binary + build cache
  deps        go mod tidy && download modules
  test        Run go test ./...
  vet         Run go vet ./...
  setup       Install system deps + build (needs sudo)
  install     Copy binary to ~/.local/bin/linboard
  stop        Stop running LinBoard process
  help        Show this help

Examples:
  ./scripts/dev.sh build
  ./scripts/dev.sh run
  make rebuild
EOF
}

cmd_build() {
  echo "==> Building $BINARY ..."
  go build -ldflags "$LDFLAGS" -o "$BINARY" "$MAIN_PKG"
  ls -lh "$BINARY"
  echo "==> Done: $ROOT/$BINARY"
}

cmd_run() {
  if [[ ! -x "$ROOT/$BINARY" ]]; then
    cmd_build
  fi
  echo "==> Running $BINARY (Super+V to open history)"
  exec "$ROOT/$BINARY" "$@"
}

cmd_clean() {
  echo "==> Removing $BINARY ..."
  rm -f "$ROOT/$BINARY"
  echo "==> Clean."
}

cmd_clean_cache() {
  echo "==> Clearing Go build cache ..."
  go clean -cache -testcache
  echo "==> Cache cleared."
}

cmd_clean_all() {
  cmd_clean
  cmd_clean_cache
}

cmd_rebuild() {
  cmd_clean
  cmd_build
}

cmd_deps() {
  echo "==> Tidying modules ..."
  go mod tidy
  go mod download
  go mod verify
  echo "==> Dependencies OK."
}

cmd_test() {
  go test ./... "$@"
}

cmd_vet() {
  go vet ./...
}

cmd_setup() {
  "$ROOT/scripts/setup.sh"
}

cmd_install() {
  cmd_build
  mkdir -p "$HOME/.local/bin"
  install -m 755 "$ROOT/$BINARY" "$HOME/.local/bin/$BINARY"
  echo "==> Installed: $HOME/.local/bin/$BINARY"
  echo "    Make sure ~/.local/bin is in your PATH"
}

cmd_stop() {
  if pgrep -x "$BINARY" >/dev/null 2>&1; then
    pkill -x "$BINARY"
    echo "==> Stopped $BINARY"
  else
    echo "==> $BINARY is not running"
  fi
}

case "${1:-help}" in
  build)       shift; cmd_build "$@" ;;
  run)         shift; cmd_run "$@" ;;
  rebuild)     shift; cmd_rebuild "$@" ;;
  clean)       shift; cmd_clean "$@" ;;
  clean-cache) shift; cmd_clean_cache "$@" ;;
  clean-all)   shift; cmd_clean_all "$@" ;;
  deps)        shift; cmd_deps "$@" ;;
  test)        shift; cmd_test "$@" ;;
  vet)         shift; cmd_vet "$@" ;;
  setup)       shift; cmd_setup "$@" ;;
  install)     shift; cmd_install "$@" ;;
  stop)        shift; cmd_stop "$@" ;;
  help|-h|--help) usage ;;
  *)
    echo "Unknown command: $1"
    echo ""
    usage
    exit 1
    ;;
esac

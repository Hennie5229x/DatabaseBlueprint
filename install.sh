#!/usr/bin/env sh
set -eu

APP_NAME="blue"
SOURCE_BINARY="./blue"
TARGET_DIR="$HOME/.local/bin"
TARGET_BINARY="$TARGET_DIR/$APP_NAME"

if [ ! -f "$SOURCE_BINARY" ]; then
  printf 'Error: binary not found at %s\n' "$SOURCE_BINARY" >&2
  exit 1
fi

mkdir -p "$TARGET_DIR"
cp "$SOURCE_BINARY" "$TARGET_BINARY"
chmod +x "$TARGET_BINARY"

printf 'Installed %s to %s\n\n' "$APP_NAME" "$TARGET_BINARY"

case ":$PATH:" in
  *":$HOME/.local/bin:"*)
    printf 'You can now run: %s\n' "$APP_NAME"
    ;;
  *)
    printf '%s is not in your PATH yet.\n' "$HOME/.local/bin"
    printf 'Add this to ~/.bashrc or ~/.zshrc:\n'
    printf 'export PATH="$HOME/.local/bin:$PATH"\n'
    ;;
esac

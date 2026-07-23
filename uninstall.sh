#!/usr/bin/env sh
set -eu

APP_NAME="blue"
TARGET_DIR="$HOME/.local/bin"
TARGET_BINARY="$TARGET_DIR/$APP_NAME"

if [ ! -e "$TARGET_BINARY" ]; then
  printf '%s is not installed at %s\n' "$APP_NAME" "$TARGET_BINARY"
  exit 0
fi

rm -f "$TARGET_BINARY"

printf 'Removed %s from %s\n' "$APP_NAME" "$TARGET_BINARY"
printf 'No shell config changes were reverted because install.sh does not modify your PATH automatically.\n'

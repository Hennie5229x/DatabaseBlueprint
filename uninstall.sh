#!/usr/bin/env sh
set -eu

APP_NAME="blue"
TARGET_DIR="${BLUE_INSTALL_DIR:-$HOME/.local/bin}"
TARGET_BINARY="$TARGET_DIR/$APP_NAME"

if [ ! -e "$TARGET_BINARY" ] && [ ! -L "$TARGET_BINARY" ]; then
    printf '%s is not installed at %s\n' "$APP_NAME" "$TARGET_BINARY"
    exit 0
fi

rm -f "$TARGET_BINARY"

printf 'Removed %s from %s\n' "$APP_NAME" "$TARGET_BINARY"
printf 'No shell configuration was changed.\n'
printf 'The installer does not modify PATH automatically.\n'

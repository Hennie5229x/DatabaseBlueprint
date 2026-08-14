#!/usr/bin/env sh
set -eu

APP_NAME="blue"
REPOSITORY="Hennie5229x/DatabaseBlueprint"
TARGET_DIR="${BLUE_INSTALL_DIR:-$HOME/.local/bin}"
TARGET_BINARY="$TARGET_DIR/$APP_NAME"
COMPLETION_DIR="${XDG_DATA_HOME:-$HOME/.local/share}/bash-completion/completions"
COMPLETION_FILE="$COMPLETION_DIR/$APP_NAME"

# Ensure curl is available.
if ! command -v curl >/dev/null 2>&1; then
    printf 'Error: curl is required to install %s.\n' "$APP_NAME" >&2
    exit 1
fi

# Detect operating system.
case "$(uname -s)" in
    Linux)
        OS="linux"
        ;;
    Darwin)
        OS="darwin"
        ;;
    *)
        printf 'Error: unsupported operating system: %s\n' "$(uname -s)" >&2
        exit 1
        ;;
esac

# Detect CPU architecture.
case "$(uname -m)" in
    x86_64 | amd64)
        ARCH="amd64"
        ;;
    arm64 | aarch64)
        ARCH="arm64"
        ;;
    *)
        printf 'Error: unsupported architecture: %s\n' "$(uname -m)" >&2
        exit 1
        ;;
esac

ASSET_NAME="${APP_NAME}-${OS}-${ARCH}"
DOWNLOAD_URL="https://github.com/${REPOSITORY}/releases/latest/download/${ASSET_NAME}"

# Only allow platforms currently built by the release workflow.
case "${OS}-${ARCH}" in
    linux-amd64 | darwin-arm64)
        ;;
    *)
        printf 'Error: no %s release is currently available for %s/%s.\n' \
            "$APP_NAME" "$OS" "$ARCH" >&2
        exit 1
        ;;
esac

TEMP_FILE="$(mktemp "${TMPDIR:-/tmp}/${APP_NAME}.XXXXXX")"

cleanup() {
    rm -f "$TEMP_FILE"
}

trap cleanup EXIT INT TERM

printf 'Downloading %s for %s/%s...\n' "$APP_NAME" "$OS" "$ARCH"

if ! curl \
    --fail \
    --location \
    --progress-bar \
    --output "$TEMP_FILE" \
    "$DOWNLOAD_URL"
then
    printf '\nError: failed to download:\n%s\n' "$DOWNLOAD_URL" >&2
    exit 1
fi

chmod +x "$TEMP_FILE"
mkdir -p "$TARGET_DIR"
mv "$TEMP_FILE" "$TARGET_BINARY"

mkdir -p "$COMPLETION_DIR"
if "$TARGET_BINARY" completion bash >"$COMPLETION_FILE"; then
    printf 'Installed bash completion to %s\n' "$COMPLETION_FILE"
else
    rm -f "$COMPLETION_FILE"
    printf 'Warning: failed to install bash completion.\n' >&2
fi

# The temporary file has been moved successfully.
trap - EXIT INT TERM

printf '\nInstalled %s to %s\n\n' "$APP_NAME" "$TARGET_BINARY"

case ":$PATH:" in
    *":$TARGET_DIR:"*)
        printf 'You can now run:\n'
        printf '  %s version\n' "$APP_NAME"
        ;;
    *)
        printf '%s is not currently in your PATH.\n\n' "$TARGET_DIR"
        printf 'Add this line to ~/.bashrc or ~/.zshrc:\n\n'
        printf '  export PATH="%s:$PATH"\n\n' "$TARGET_DIR"
        printf 'Then open a new terminal and run:\n'
        printf '  %s version\n' "$APP_NAME"
        ;;
esac

if [ "$OS" = "darwin" ]; then
    printf '\nIf macOS blocks the unsigned binary, run:\n'
    printf '  xattr -d com.apple.quarantine "%s"\n' "$TARGET_BINARY"
fi

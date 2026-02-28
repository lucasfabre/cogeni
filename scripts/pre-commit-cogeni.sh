#!/bin/bash
set -e

# Use the first argument as the cogeni command if provided, otherwise default to 'cogeni'
COGENI_CMD=${1:-cogeni}
shift || true

# Check if cogeni is installed or available at the specified path
if ! command -v "$COGENI_CMD" &>/dev/null && [ ! -f "$COGENI_CMD" ]; then
	echo "Error: $COGENI_CMD is not installed or not in PATH."
	echo "Please build cogeni or ensure it is in your PATH to run this pre-commit hook."
	exit 1
fi

# Run cogeni
# Note: pre-commit will automatically detect if files were modified.
"$COGENI_CMD" "$@"

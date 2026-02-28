#!/bin/bash
set -e

# Use the first argument as the cogeni command if provided, otherwise default to 'cogeni'
COGENI_CMD=${1:-cogeni}
shift || true

# Check if cogeni is installed or available at the specified path
if ! command -v "$COGENI_CMD" &>/dev/null && [ ! -f "$COGENI_CMD" ]; then
	echo "Error: $COGENI_CMD is not installed or not in PATH."
	echo "Please install cogeni to run this pre-commit hook."
	exit 1
fi

# Run cogeni
# Using "$@" allows passing remaining arguments from the hook configuration
"$COGENI_CMD" run "$@"

# Check if there are any changes in the git index or working directory
# pre-commit usually stashes unstaged changes, so this diff reflects changes made by cogeni
if ! git diff --quiet; then
	echo "Files were modified by cogeni. Please stage the changes and commit again."
	exit 1
fi

exit 0

#!/bin/bash
set -e

# Check if cogeni is installed
if ! command -v cogeni &>/dev/null; then
	echo "Error: cogeni is not installed or not in PATH."
	echo "Please install cogeni to run this pre-commit hook."
	exit 1
fi

# Run cogeni
# Using "$@" allows passing arguments from the hook configuration if needed
cogeni run "$@"

# Check if there are any changes in the git index or working directory
# pre-commit usually stashes unstaged changes, so this diff reflects changes made by cogeni
if ! git diff --quiet; then
	echo "Files were modified by cogeni. Please stage the changes and commit again."
	exit 1
fi

exit 0

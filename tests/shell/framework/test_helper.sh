#!/usr/bin/env bash

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

PASSES=0
FAILURES=0

# Determine absolute path to cogeni binary if not already set
if [ -z "$COGENI_BIN" ]; then
	if [ -f "./cogeni.exe" ]; then
		COGENI_BIN="$(realpath ./cogeni.exe)"
		export COGENI_BIN
	elif [ -f "./cogeni" ]; then
		COGENI_BIN="$(realpath ./cogeni)"
		export COGENI_BIN
	else
		# Fallback - assume we are in project root
		COGENI_BIN="$(pwd)/cogeni"
		export COGENI_BIN
	fi
fi

describe() {
	echo -e "
$1"
}

it() {
	local name="$1"
	local cmd="$2"
	local helper_path
	helper_path="$(realpath "${BASH_SOURCE[0]}")"

	if [ "$(type -t setup)" == "function" ]; then
		setup
	fi

	# Run the command with COGENI_BIN exported
	# Use a subshell with 'set -e' to ensure any failing command (including assertions)
	# results in a non-zero exit code for the entire test block.
	# We source the helper_path to make assertion functions available.
	if bash -c "set -e; export COGENI_BIN=\"$COGENI_BIN\"; source \"$helper_path\"; $cmd"; then
		echo -e "  ${GREEN}[PASS]${NC} $name"
		PASSES=$((PASSES + 1))
	else
		echo -e "  ${RED}[FAIL]${NC} $name"
		FAILURES=$((FAILURES + 1))
	fi

	if [ "$(type -t teardown)" == "function" ]; then
		teardown
	fi
}

assert_contains() {
	local haystack="$1"
	local needle="$2"
	if [[ $haystack == *"$needle"* ]]; then
		return 0
	else
		echo "Expected to contain '$needle' but got '$haystack'" >&2
		return 1
	fi
}

assert_not_contains() {
	local haystack="$1"
	local needle="$2"
	if [[ $haystack != *"$needle"* ]]; then
		return 0
	else
		return 1
	fi
}

assert_eq() {
	local actual="$1"
	local expected="$2"
	if [[ $actual == "$expected" ]]; then
		return 0
	else
		echo "Expected '$expected' but got '$actual'" >&2
		return 1
	fi
}

assert_file_exists() {
	if [ -f "$1" ]; then
		return 0
	else
		echo "File '$1' does not exist" >&2
		return 1
	fi
}

summary() {
	echo -e "
Shell Tests complete: $PASSES passed, $FAILURES failed"
}

exit_code() {
	if [ $FAILURES -gt 0 ]; then
		exit 1
	else
		exit 0
	fi
}

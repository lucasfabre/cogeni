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

# Set shared grammar location for tests to avoid re-downloading/re-compiling
if [ -z "$COGENI_GRAMMAR_LOCATION" ]; then
	# Try to find project root via git, fallback to pwd (assuming we are in root)
	PROJECT_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
	export COGENI_GRAMMAR_LOCATION="$PROJECT_ROOT/.cache/test-grammars"
	mkdir -p "$COGENI_GRAMMAR_LOCATION"
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

	# Record start time (using nanoseconds if available)
	local start_time
	start_time=$(date +%s%N)

	# Run the command with COGENI_BIN exported
	# Use a subshell with 'set -e' to ensure any failing command (including assertions)
	# results in a non-zero exit code for the entire test block.
	# We source the helper_path to make assertion functions available.
	if bash -c "set -e; export COGENI_BIN=\"$COGENI_BIN\"; export COGENI_GRAMMAR_LOCATION=\"$COGENI_GRAMMAR_LOCATION\"; source \"$helper_path\"; $cmd"; then
		local end_time
		end_time=$(date +%s%N)
		local elapsed_ns=$((end_time - start_time))
		# Convert ns to seconds with 3 decimal places using python3
		local elapsed_s
		elapsed_s=$(python3 -c "print(f'{($elapsed_ns / 1000000000):.3f}')")
		echo -e "  ${GREEN}[PASS]${NC} $name (${elapsed_s}s)"
		PASSES=$((PASSES + 1))
	else
		local end_time
		end_time=$(date +%s%N)
		local elapsed_ns=$((end_time - start_time))
		local elapsed_s
		elapsed_s=$(python3 -c "print(f'{($elapsed_ns / 1000000000):.3f}')")
		echo -e "  ${RED}[FAIL]${NC} $name (${elapsed_s}s)"
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

wait_for_log() {
	local log_file="$1"
	local pattern="$2"
	local timeout=50 # 5 seconds (0.1s * 50)
	local i=0
	while [ $i -lt $timeout ]; do
		if grep -q "$pattern" "$log_file"; then
			return 0
		fi
		sleep 0.1
		i=$((i + 1))
	done
	return 1
}

wait_for_file_content() {
	local file="$1"
	local pattern="$2"
	local timeout=50
	local i=0
	while [ $i -lt $timeout ]; do
		if [ -f "$file" ] && grep -q "$pattern" "$file"; then
			return 0
		fi
		sleep 0.1
		i=$((i + 1))
	done
	return 1
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

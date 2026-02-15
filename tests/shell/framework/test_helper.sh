#!/usr/bin/env bash

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

PASSES=0
FAILURES=0

describe() {
	echo -e "
$1"
}

it() {
	local name="$1"
	local cmd="$2"

	if eval "$cmd"; then
		echo -e "  ${GREEN}[PASS]${NC} $name"
		PASSES=$((PASSES + 1))
	else
		echo -e "  ${RED}[FAIL]${NC} $name"
		FAILURES=$((FAILURES + 1))
	fi
}

assert_contains() {
	local haystack="$1"
	local needle="$2"
	if [[ $haystack == *"$needle"* ]]; then
		return 0
	else
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

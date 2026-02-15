#!/usr/bin/env bash
set -e
# ... rest of the file

set -e

# ANSI color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "Starting Test Suite..."

# Source the shell test framework helper
# shellcheck source=/dev/null
source tests/shell/framework/test_helper.sh

# Run all shell tests
for test_file in tests/shell/*_test.sh; do
	echo -e "\nRunning $(basename "$test_file")..."
	# shellcheck source=/dev/null
	source "$test_file"
done

# Output shell test summary and exit if failures occurred
summary
if [ "$FAILURES" -gt 0 ]; then
	echo -e "${RED}[FAIL]${NC} Shell tests failed."
	exit 1
fi

echo -e "\nRunning Lua API tests..."
# Execute Lua tests via the CLI
if ! $COGENI_BIN run tests/lua/framework/test.lua; then
	echo -e "${RED}[FAIL]${NC} Lua tests failed"
	exit 1
fi

echo -e "\n${GREEN}All tests passed successfully!${NC}"
exit 0

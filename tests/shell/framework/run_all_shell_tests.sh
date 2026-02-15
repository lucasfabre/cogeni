#!/usr/bin/env bash

set -e

# Source the test helper
# shellcheck source=/dev/null
source tests/shell/framework/test_helper.sh

# Find and source all shell test files
for test_file in tests/shell/*_test.sh; do
	echo -e "\nRunning $(basename "$test_file")..."
	# shellcheck source=/dev/null
	source "$test_file"
done

# Output final summary and exit with appropriate code
summary
exit_code

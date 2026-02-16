#!/usr/bin/env bash
# shellcheck disable=SC2016

describe "Watch Command"

setup() {
	TEST_DIR="temp_watch_test"
	mkdir -p "$TEST_DIR"
	cd "$TEST_DIR" || exit 1
}

teardown() {
	cd ..
	rm -rf "$TEST_DIR"
	# Kill any lingering watch processes if variable is set
	if [ -n "$WATCH_PID" ]; then
		kill "$WATCH_PID" 2>/dev/null || true
	fi
}

it "should watch a simple file and update on change" '
    echo "write(\"out\", \"version 1\")" > entry.lua
    echo "cogeni.outfile(\"out\", \"output.txt\")" >> entry.lua

    # Start watch
    $COGENI_BIN watch entry.lua > watch.log 2>&1 &
    WATCH_PID=$!

    # Wait for initial build
    sleep 2

    assert_file_exists "output.txt"
    assert_contains "$(cat output.txt)" "version 1"

    # Modify file
    echo "write(\"out\", \"version 2\")" > entry.lua
    echo "cogeni.outfile(\"out\", \"output.txt\")" >> entry.lua

    sleep 2

    assert_contains "$(cat output.txt)" "version 2"
'

it "should watch dependencies and update on change" '
    # Create dependency
    echo "-- <cogeni>" > dep.lua
    echo "write(\"dep\", \"dep 1\")" >> dep.lua
    echo "cogeni.outfile(\"dep\", \"dep_output.txt\")" >> dep.lua
    echo "-- </cogeni>" >> dep.lua

    # Entry point that uses dependency
    echo "cogeni.process(\"dep.lua\")" > entry.lua

    # Start watch
    $COGENI_BIN watch entry.lua > watch.log 2>&1 &
    WATCH_PID=$!

    sleep 2

    assert_file_exists "dep_output.txt"
    assert_contains "$(cat dep_output.txt)" "dep 1"

    # Modify dependency
    echo "-- <cogeni>" > dep.lua
    echo "write(\"dep\", \"dep 2\")" >> dep.lua
    echo "cogeni.outfile(\"dep\", \"dep_output.txt\")" >> dep.lua
    echo "-- </cogeni>" >> dep.lua

    sleep 2

    assert_contains "$(cat dep_output.txt)" "dep 2"
'

it "should handle atomic saves (Rename)" '
    echo "write(\"out\", \"version 1\")" > entry.lua
    echo "cogeni.outfile(\"out\", \"output.txt\")" >> entry.lua

    $COGENI_BIN watch entry.lua > watch.log 2>&1 &
    WATCH_PID=$!

    sleep 2
    assert_contains "$(cat output.txt)" "version 1"

    # Atomic save simulation: create tmp and move
    echo "write(\"out\", \"version 2\")" > entry.lua.tmp
    echo "cogeni.outfile(\"out\", \"output.txt\")" >> entry.lua.tmp
    mv entry.lua.tmp entry.lua

    sleep 2

    assert_contains "$(cat output.txt)" "version 2"
'

it "should detect circular dependencies" '
    # Create file A
    echo "-- <cogeni>" > A.lua
    echo "cogeni.process(\"B.lua\")" >> A.lua
    echo "-- </cogeni>" >> A.lua

    # Create file B
    echo "-- <cogeni>" > B.lua
    echo "cogeni.process(\"A.lua\")" >> B.lua
    echo "-- </cogeni>" >> B.lua

    # Start watch
    $COGENI_BIN watch A.lua > watch.log 2>&1 &
    WATCH_PID=$!

    sleep 2

    assert_contains "$(cat watch.log)" "circular dependency detected"
'

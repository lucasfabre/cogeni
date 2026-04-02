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
	local pid_file="$TEST_DIR/.watch.pid"
	# The tests run in a subshell, so persist the watcher PID to disk for teardown.
	if [ -f "$pid_file" ]; then
		WATCH_PID="$(cat "$pid_file")"
		if [ -n "$WATCH_PID" ]; then
			if command -v cmd.exe >/dev/null 2>&1; then
				cmd.exe /c taskkill //PID "$WATCH_PID" //T //F >/dev/null 2>&1 || true
			else
				kill "$WATCH_PID" 2>/dev/null || true
				wait "$WATCH_PID" 2>/dev/null || true
			fi
			sleep 0.5
		fi
		rm -f "$pid_file"
	fi
	for _ in 1 2 3 4 5; do
		if rm -rf "$TEST_DIR" 2>/dev/null; then
			return
		fi
		sleep 0.2
	done
	rm -rf "$TEST_DIR"
}

it "should watch a simple file and update on change" '
    echo "write(\"out\", \"version 1\")" > entry.lua
    echo "cogeni.outfile(\"out\", \"output.txt\")" >> entry.lua

    # Start watch
    $COGENI_BIN watch entry.lua > watch.log 2>&1 &
    WATCH_PID=$!
    echo "$WATCH_PID" > .watch.pid

    # Wait for initial build
    wait_for_log "watch.log" "Watching for changes" || { echo "Timeout waiting for start"; cat watch.log; exit 1; }

    wait_for_file_content "output.txt" "version 1" || { echo "Timeout waiting for version 1"; cat watch.log; exit 1; }

    # Modify file
    echo "write(\"out\", \"version 2\")" > entry.lua
    echo "cogeni.outfile(\"out\", \"output.txt\")" >> entry.lua

    wait_for_file_content "output.txt" "version 2" || { echo "Timeout waiting for version 2"; cat watch.log; exit 1; }
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
    echo "$WATCH_PID" > .watch.pid

    wait_for_log "watch.log" "Watching for changes" || { echo "Timeout waiting for start"; cat watch.log; exit 1; }

    wait_for_file_content "dep_output.txt" "dep 1" || { echo "Timeout waiting for dep 1"; cat watch.log; exit 1; }

    # Modify dependency
    echo "-- <cogeni>" > dep.lua
    echo "write(\"dep\", \"dep 2\")" >> dep.lua
    echo "cogeni.outfile(\"dep\", \"dep_output.txt\")" >> dep.lua
    echo "-- </cogeni>" >> dep.lua

    wait_for_file_content "dep_output.txt" "dep 2" || { echo "Timeout waiting for dep 2"; cat watch.log; exit 1; }
'

it "should handle atomic saves (Rename)" '
    echo "write(\"out\", \"version 1\")" > entry.lua
    echo "cogeni.outfile(\"out\", \"output.txt\")" >> entry.lua

    $COGENI_BIN watch entry.lua > watch.log 2>&1 &
    WATCH_PID=$!
    echo "$WATCH_PID" > .watch.pid

    wait_for_log "watch.log" "Watching for changes" || { echo "Timeout waiting for start"; cat watch.log; exit 1; }
    wait_for_file_content "output.txt" "version 1" || { echo "Timeout waiting for version 1"; cat watch.log; exit 1; }

    # Atomic save simulation: create tmp and move
    echo "write(\"out\", \"version 2\")" > entry.lua.tmp
    echo "cogeni.outfile(\"out\", \"output.txt\")" >> entry.lua.tmp
    mv entry.lua.tmp entry.lua

    wait_for_file_content "output.txt" "version 2" || { echo "Timeout waiting for version 2"; cat watch.log; exit 1; }
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
    echo "$WATCH_PID" > .watch.pid

    # Wait for cycle detection message
    # "circular dependency detected" might be part of the error message
    wait_for_log "watch.log" "circular dependency detected" || { echo "Timeout waiting for cycle detection"; cat watch.log; exit 1; }
'

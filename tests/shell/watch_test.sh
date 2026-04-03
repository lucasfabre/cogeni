#!/usr/bin/env bash
# shellcheck disable=SC2016

describe "Watch Command"

setup() {
	TEST_DIR="temp_watch_test"
	mkdir -p "$TEST_DIR"
	cd "$TEST_DIR" || exit 1
}

is_windows() {
	[ -n "$WINDIR" ] || command -v cmd.exe >/dev/null 2>&1
}

kill_watch_processes() {
	local pid_file="$TEST_DIR/.watch.pid"
	if is_windows; then
		powershell.exe -NoProfile -Command "Get-Process -Name cogeni -ErrorAction SilentlyContinue | Stop-Process -Force" >/dev/null 2>&1 || true
		return
	fi

	if [ -f "$pid_file" ]; then
		WATCH_PID="$(cat "$pid_file")"
		if [ -n "$WATCH_PID" ]; then
			kill "$WATCH_PID" 2>/dev/null || true
			wait "$WATCH_PID" 2>/dev/null || true
		fi
	fi
}

watch_processes_gone() {
	if is_windows; then
		powershell.exe -NoProfile -Command "if (Get-Process -Name cogeni -ErrorAction SilentlyContinue) { exit 1 }" >/dev/null 2>&1
		return $?
	fi

	local pid_file="$TEST_DIR/.watch.pid"
	if [ ! -f "$pid_file" ]; then
		return 0
	fi

	WATCH_PID="$(cat "$pid_file")"
	if [ -z "$WATCH_PID" ]; then
		return 0
	fi

	if kill -0 "$WATCH_PID" 2>/dev/null; then
		return 1
	fi
	return 0
}

print_watch_diagnostics() {
	if is_windows; then
		echo "Remaining cogeni processes on Windows:" >&2
		powershell.exe -NoProfile -Command "Get-Process -Name cogeni -ErrorAction SilentlyContinue | Format-Table -AutoSize Id,ProcessName,Path" >&2 || true
	fi
	if [ -f "$TEST_DIR/watch.log" ]; then
		echo "watch.log:" >&2
		cat "$TEST_DIR/watch.log" >&2
	fi
}

teardown() {
	cd ..
	kill_watch_processes
	for _ in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15; do
		if watch_processes_gone && rm -rf "$TEST_DIR" 2>/dev/null; then
			return
		fi
		sleep 0.4
	done
	print_watch_diagnostics
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

it "should detect circular dependencies and format them clearly" '
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
	    wait_for_log "watch.log" "circular dependency detected" || { echo "Timeout waiting for cycle detection"; cat watch.log; exit 1; }
	    wait_for_log "watch.log" "A.lua" || { echo "Timeout waiting for cycle detection formatting"; cat watch.log; exit 1; }

	    # Ensure watch process stays alive after a failure
	    kill -0 $WATCH_PID || { echo "Watch process died after cycle detection"; exit 1; }
'

it "should debounce rapid successive edits" '
    echo "write(\"out\", \"version 1\")" > entry.lua
    echo "cogeni.outfile(\"out\", \"output.txt\")" >> entry.lua

    # Start watch
    $COGENI_BIN watch entry.lua > watch.log 2>&1 &
    WATCH_PID=$!
    echo "$WATCH_PID" > .watch.pid

    wait_for_log "watch.log" "Watching for changes" || { echo "Timeout waiting for start"; cat watch.log; exit 1; }
    wait_for_file_content "output.txt" "version 1" || { echo "Timeout waiting for version 1"; cat watch.log; exit 1; }

    # Perform rapid edits
    echo "write(\"out\", \"version 2\")" > entry.lua
    echo "cogeni.outfile(\"out\", \"output.txt\")" >> entry.lua
    sleep 0.05
    echo "write(\"out\", \"version 3\")" > entry.lua
    echo "cogeni.outfile(\"out\", \"output.txt\")" >> entry.lua
    sleep 0.05
    echo "write(\"out\", \"version 4\")" > entry.lua
    echo "cogeni.outfile(\"out\", \"output.txt\")" >> entry.lua

    wait_for_file_content "output.txt" "version 4" || { echo "Timeout waiting for version 4"; cat watch.log; exit 1; }

    # Verify that only 2 builds happened (initial + 1 debounced rebuild)
    BUILD_COUNT=$(grep -c "Starting build..." watch.log)
    if [ "$BUILD_COUNT" -gt 2 ]; then
        echo "Expected at most 2 builds due to debouncing, but saw $BUILD_COUNT"
        cat watch.log
        exit 1
    fi
'

it "should recover from script execution errors" '
    echo "write(\"out\", \"version 1\")" > entry.lua
    echo "cogeni.outfile(\"out\", \"output.txt\")" >> entry.lua

    # Start watch
    $COGENI_BIN watch entry.lua > watch.log 2>&1 &
    WATCH_PID=$!
    echo "$WATCH_PID" > .watch.pid

    wait_for_log "watch.log" "Watching for changes" || { echo "Timeout waiting for start"; cat watch.log; exit 1; }
    wait_for_file_content "output.txt" "version 1" || { echo "Timeout waiting for version 1"; cat watch.log; exit 1; }

    # Introduce a syntax error
    echo "INVALID_SYNTAX(" > entry.lua

    wait_for_log "watch.log" "Execution failed in" || { echo "Timeout waiting for error log"; cat watch.log; exit 1; }

    # Watcher should still be alive
    kill -0 $WATCH_PID || { echo "Watch process died after script error"; exit 1; }

    # Fix the error
    echo "write(\"out\", \"version 2\")" > entry.lua
    echo "cogeni.outfile(\"out\", \"output.txt\")" >> entry.lua

    wait_for_file_content "output.txt" "version 2" || { echo "Timeout waiting for version 2"; cat watch.log; exit 1; }
'

it "should recover when the initial build fails" '
    echo "INVALID_SYNTAX(" > entry.lua

    $COGENI_BIN watch entry.lua > watch.log 2>&1 &
    WATCH_PID=$!
    echo "$WATCH_PID" > .watch.pid

    wait_for_log "watch.log" "Execution failed in" || { echo "Timeout waiting for startup error"; cat watch.log; exit 1; }

    # Watcher should still be alive and ready to react to the fix
    kill -0 $WATCH_PID || { echo "Watch process died after initial failure"; exit 1; }

    echo "write(\"out\", \"version 1\")" > entry.lua
    echo "cogeni.outfile(\"out\", \"output.txt\")" >> entry.lua

    wait_for_file_content "output.txt" "version 1" || { echo "Timeout waiting for recovery rebuild"; cat watch.log; exit 1; }
'

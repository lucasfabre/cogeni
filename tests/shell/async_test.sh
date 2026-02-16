#!/usr/bin/env bash
# shellcheck disable=SC2016

describe "Async Support"

it "should support multiple concurrent async tasks" '
    cat > test_multi_async.lua <<EOF
        local results = {}
        for i=1,5 do
            async(function()
                await(sleep, 0.1 * (6-i))
                results[i] = i
                print("FINISH:" .. i)
            end)
        end
EOF
    out=$($COGENI_BIN run test_multi_async.lua)
    for i in {1..5}; do
        assert_contains "$out" "FINISH:$i"
    done
    rm test_multi_async.lua
'

it "should handle errors in async tasks" '
    cat > test_async_err.lua <<EOF
        async(function()
            local status, err = pcall(function()
                error("something went wrong")
            end)
            if not status then
                print("ASYNC_ERROR: " .. err)
            end
        end)
EOF
    # Errors in async tasks currently might just print to stderr and not fail the main process immediately
    # We should verify how it behaves.
    out=$($COGENI_BIN run test_async_err.lua 2>&1)
    assert_contains "$out" "something went wrong"
    rm test_async_err.lua
'

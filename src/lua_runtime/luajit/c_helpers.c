#include "_cgo_export.h"
#include <lua.h>
#include <lauxlib.h>

/*
  Trampoline function that is pushed to Lua stack as a C closure.
  It retrieves the Go function handle from the upvalue and calls into Go.

  The Go function returns an integer status code:
  >= 0: Number of results to return (normal return)
  -1:   Yield (requires stack preparation in Go: arguments/results)
  -2:   Error (message on top of stack)
*/
int lua_trampoline(lua_State *L) {
    // Get the handle from upvalue (index 1)
    int handle = (int)lua_tointeger(L, lua_upvalueindex(1));

    // Call the exported Go function
    int status = InvokeGoCallback(L, handle);

    if (status == -1) {
        // Yield: Go code has pushed args/results or whatever is needed.
        // We yield with the number of results currently on stack?
        // Usually yield takes num_results.
        // But how do we pass num_results from Go?
        // InvokeGoCallback returns status.
        // We need a side channel or a specific convention.
        // Let's assume for Yield(n), Go returns (-100 - n).
        // No, simple yield(0) is most common.
        // Let's use a global/thread-local variable? No.

        // Let's change the protocol:
        // status >= 0: number of results (OK)
        // status < -1000: Error?

        // Simpler:
        // Go returns:
        // N >= 0: return N results
        // -1: Yield 0 results
        // -2: Error (msg on stack)
        // -(10 + n): Yield n results?

        return lua_yield(L, 0);
    }

    if (status == -2) {
        return lua_error(L);
    }

    // If status < -10, it's a yield with results
    if (status <= -10) {
        int nresults = -(status + 10);
        return lua_yield(L, nresults);
    }

    return status;
}

// Wrapper for luaL_error to avoid variadic issues in CGo
int lua_error_wrapper(lua_State *L, const char *msg) {
    return luaL_error(L, "%s", msg);
}

// Safe call wrapper
int lua_pcall_wrapper(lua_State *L, int nargs, int nresults) {
    return lua_pcall(L, nargs, nresults, 0);
}

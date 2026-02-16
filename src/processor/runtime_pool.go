package processor

import (
	"github.com/lucasfabre/codegen/src/config"
	luaruntime "github.com/lucasfabre/codegen/src/lua_runtime"
)

// RuntimePool manages a pool of reusable LuaRuntime instances.
// This reduces the overhead of creating new Lua states for each file.
type RuntimePool struct {
	cfg     *config.Config
	pool    chan *luaruntime.LuaRuntime
	maxSize int
}

// NewRuntimePool creates a new runtime pool with the specified maximum size.
func NewRuntimePool(cfg *config.Config, maxSize int) *RuntimePool {
	return &RuntimePool{
		cfg:     cfg,
		pool:    make(chan *luaruntime.LuaRuntime, maxSize),
		maxSize: maxSize,
	}
}

// Acquire gets a runtime from the pool or creates a new one if needed.
func (rp *RuntimePool) Acquire() (*luaruntime.LuaRuntime, error) {
	select {
	case rt := <-rp.pool:
		// Return a runtime from the pool
		return rt, nil
	default:
		// Pool is empty, create a new runtime
		return luaruntime.New(rp.cfg)
	}
}

// Release returns a runtime to the pool for reuse.
func (rp *RuntimePool) Release(rt *luaruntime.LuaRuntime) {
	// always close the runtime to avoid issues with glua-async state on reuse.
	rt.Close()
}

// Close shuts down all runtimes in the pool.
func (rp *RuntimePool) Close() {
	close(rp.pool)
	for rt := range rp.pool {
		rt.Close()
	}
}

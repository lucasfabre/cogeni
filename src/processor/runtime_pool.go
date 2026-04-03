package processor

import (
	"github.com/lucasfabre/cogeni/src/config"
	luaruntime "github.com/lucasfabre/cogeni/src/lua_runtime"
)

// RuntimePool manages a pool of reusable LuaRuntime instances.
// This reduces the overhead of creating new Lua states for each file.
type RuntimePool struct {
	cfg     *config.Config
	sem     chan struct{}
	maxSize int
}

// NewRuntimePool creates a new runtime pool with the specified maximum size.
func NewRuntimePool(cfg *config.Config, maxSize int) *RuntimePool {
	return &RuntimePool{
		cfg:     cfg,
		sem:     make(chan struct{}, maxSize),
		maxSize: maxSize,
	}
}

// Acquire limits concurrency and returns a new runtime.
func (rp *RuntimePool) Acquire() (*luaruntime.LuaRuntime, error) {
	rp.sem <- struct{}{} // Block until a slot is available
	rt, err := luaruntime.New(rp.cfg)
	if err != nil {
		<-rp.sem // Free the slot if initialization fails
		return nil, err
	}
	return rt, nil
}

// Release closes the runtime and frees up a concurrency slot.
func (rp *RuntimePool) Release(rt *luaruntime.LuaRuntime) {
	// always close the runtime to avoid issues with glua-async state on reuse.
	if rt != nil {
		rt.Close()
	}
	<-rp.sem // Free the slot
}

// Close ensures the pool is in a valid state on teardown.
func (rp *RuntimePool) Close() {
	// no-op for now, runtimes are closed on Release
}

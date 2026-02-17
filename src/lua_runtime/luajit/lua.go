package luajit

/*
#cgo pkg-config: luajit
#include <lua.h>
#include <lualib.h>
#include <lauxlib.h>
#include <stdlib.h>
#include <stdio.h>

// Forward declaration of the C trampoline function defined in c_helpers.c
int lua_trampoline(lua_State *L);
int lua_error_wrapper(lua_State *L, const char *msg);
int lua_pcall_wrapper(lua_State *L, int nargs, int nresults);
*/
import "C"
import (
	"fmt"
	"runtime/cgo"
	"sync"
	"unsafe"
)

// Lua Status Codes
const (
	OK                = 0
	YIELD             = 1
	ERRRUN            = 2
	ERRSYNTAX         = 3
	ERRMEM            = 4
	ERRERR            = 5
	LUA_REGISTRYINDEX = C.LUA_REGISTRYINDEX
	MULTRET           = C.LUA_MULTRET
)

// Status conventions for Go callbacks
const (
	GO_YIELD_0 = -1
	GO_ERROR   = -2
)

// LuaFunction is the signature for Go functions callable from Lua
// Returns: number of results, OR GO_YIELD_0, OR GO_ERROR
type LuaFunction func(*State) int

//export InvokeGoCallback
func InvokeGoCallback(L *C.lua_State, handle C.int) (ret C.int) {
	// Retrieve State from Lua Registry
	stateKey := C.CString("_GO_STATE")
	defer C.free(unsafe.Pointer(stateKey))

	C.lua_getfield(L, C.LUA_REGISTRYINDEX, stateKey)
	p := C.lua_touserdata(L, -1)
	C.lua_settop(L, -2)

	if p == nil {
		errMsg := C.CString("go state not found in registry")
		defer C.free(unsafe.Pointer(errMsg))
		C.lua_pushstring(L, errMsg)
		return GO_ERROR
	}

	h := cgo.Handle(uintptr(unsafe.Pointer(p)))
	s := h.Value().(*State)

	s.mu.RLock()
	fn, ok := s.registry[int(handle)]
	s.mu.RUnlock()

	if !ok {
		errMsg := C.CString("go callback not found")
		defer C.free(unsafe.Pointer(errMsg))
		C.lua_pushstring(L, errMsg)
		return GO_ERROR
	}

	currentState := &State{
		L:          L,
		registry:   s.registry,
		mu:         s.mu,
		nextHandle: s.nextHandle,
	}

	// Safely recover from panics
	defer func() {
		if r := recover(); r != nil {
			errMsg := C.CString(fmt.Sprintf("go panic: %v", r))
			defer C.free(unsafe.Pointer(errMsg))
			C.lua_pushstring(L, errMsg)
			ret = GO_ERROR
		}
	}()

	ret = C.int(fn(currentState))
	return
}

// State wraps the C lua_State and shared Go context
type State struct {
	L          *C.lua_State
	registry   map[int]LuaFunction
	mu         *sync.RWMutex
	nextHandle *int
	handle     cgo.Handle
}

// NewState creates a new Lua state
func NewState() *State {
	L := C.luaL_newstate()
	if L == nil {
		return nil
	}

	s := &State{
		L:          L,
		registry:   make(map[int]LuaFunction),
		mu:         &sync.RWMutex{},
		nextHandle: new(int),
	}
	*s.nextHandle = 1

	// Create handle to keep s alive and safe for C access
	h := cgo.NewHandle(s)
	s.handle = h

	key := C.CString("_GO_STATE")
	defer C.free(unsafe.Pointer(key))
	// Store handle as lightuserdata (safe because it's uintptr)
	C.lua_pushlightuserdata(L, unsafe.Pointer(uintptr(h)))
	C.lua_setfield(L, C.LUA_REGISTRYINDEX, key)

	return s
}

// Close closes the Lua state
func (s *State) Close() {
	if s.L != nil {
		C.fflush(C.stdout)
		C.fflush(C.stderr)
		// Clean up handle if it's the main state/owner
		if s.handle > 0 {
			s.handle.Delete()
			s.handle = 0
		}
		C.lua_close(s.L)
		s.L = nil
	}
}

// OpenLibs opens standard Lua libraries
func (s *State) OpenLibs() {
	C.luaL_openlibs(s.L)
}

// LoadFile loads a file as a chunk
func (s *State) LoadFile(path string) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	if status := C.luaL_loadfile(s.L, cPath); status != 0 {
		return s.errorFromStack(status)
	}
	return nil
}

// LoadString loads a string as a chunk without executing it
func (s *State) LoadString(str string) error {
	cStr := C.CString(str)
	defer C.free(unsafe.Pointer(cStr))

	if status := C.luaL_loadstring(s.L, cStr); status != 0 {
		return s.errorFromStack(status)
	}
	return nil
}

// DoString executes a string
func (s *State) DoString(str string) error {
	if err := s.LoadString(str); err != nil {
		return err
	}

	// Execute the chunk
	if status := C.lua_pcall(s.L, 0, C.LUA_MULTRET, 0); status != 0 {
		return s.errorFromStack(status)
	}
	return nil
}

// Call calls a function safely (pcall).
func (s *State) Call(nargs, nresults int) {
	// Use protected call
	status := C.lua_pcall_wrapper(s.L, C.int(nargs), C.int(nresults))
	if status != 0 {
		msg := s.CheckString(-1)
		s.Pop(1)
		panic(fmt.Sprintf("lua error: %s", msg))
	}
}

// DoFile executes a file
func (s *State) DoFile(path string) error {
	if err := s.LoadFile(path); err != nil {
		return err
	}
	// Execute the chunk
	if status := C.lua_pcall(s.L, 0, C.LUA_MULTRET, 0); status != 0 {
		return s.errorFromStack(status)
	}
	return nil
}

// errorFromStack pops the error message from the stack and returns it as a Go error
func (s *State) errorFromStack(status C.int) error {
	msg := C.GoString(C.lua_tolstring(s.L, -1, nil))
	C.lua_settop(s.L, -2) // pop error message
	return fmt.Errorf("lua error (%d): %s", status, msg)
}

// register stores the function and returns a handle
func (s *State) register(fn LuaFunction) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	h := *s.nextHandle
	*s.nextHandle++
	s.registry[h] = fn
	return h
}

// unregister removes the function
func (s *State) unregister(h int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.registry, h)
}

// PushGoFunction pushes a Go function as a C closure with an upvalue
func (s *State) PushGoFunction(fn LuaFunction) {
	handle := s.register(fn)
	C.lua_pushinteger(s.L, C.lua_Integer(handle))
	C.lua_pushcclosure(s.L, (C.lua_CFunction)(C.lua_trampoline), 1)
}

// --- Stack Operations ---

func (s *State) PushNumber(n float64) {
	C.lua_pushnumber(s.L, C.lua_Number(n))
}

func (s *State) PushString(str string) {
	cStr := C.CString(str)
	defer C.free(unsafe.Pointer(cStr))
	C.lua_pushstring(s.L, cStr)
}

func (s *State) PushBool(b bool) {
	val := C.int(0)
	if b {
		val = 1
	}
	C.lua_pushboolean(s.L, val)
}

func (s *State) PushNil() {
	C.lua_pushnil(s.L)
}

func (s *State) PushThread() int {
	return int(C.lua_pushthread(s.L))
}

func (s *State) PushValue(idx int) {
	C.lua_pushvalue(s.L, C.int(idx))
}

func (s *State) GetTop() int {
	return int(C.lua_gettop(s.L))
}

func (s *State) SetTop(idx int) {
	C.lua_settop(s.L, C.int(idx))
}

func (s *State) Pop(n int) {
	C.lua_settop(s.L, C.int(-n-1))
}

func (s *State) CheckString(idx int) string {
	var len C.size_t
	cStr := C.lua_tolstring(s.L, C.int(idx), &len)
	if cStr == nil {
		panic(fmt.Sprintf("bad argument #%d (string expected, got %s)", idx, s.TypeName(idx)))
	}
	return C.GoStringN(cStr, C.int(len))
}

func (s *State) CheckNumber(idx int) float64 {
	if C.lua_isnumber(s.L, C.int(idx)) == 0 {
		panic(fmt.Sprintf("bad argument #%d (number expected, got %s)", idx, s.TypeName(idx)))
	}
	return float64(C.lua_tonumber(s.L, C.int(idx)))
}

func (s *State) TypeName(idx int) string {
	t := C.lua_type(s.L, C.int(idx))
	return C.GoString(C.lua_typename(s.L, t))
}

func (s *State) IsString(idx int) bool {
	return C.lua_type(s.L, C.int(idx)) == C.LUA_TSTRING
}

func (s *State) IsNumber(idx int) bool {
	return C.lua_type(s.L, C.int(idx)) == C.LUA_TNUMBER
}

func (s *State) IsBool(idx int) bool {
	return C.lua_type(s.L, C.int(idx)) == C.LUA_TBOOLEAN
}

func (s *State) IsFunction(idx int) bool {
	return C.lua_type(s.L, C.int(idx)) == C.LUA_TFUNCTION
}

func (s *State) IsThread(idx int) bool {
	return C.lua_type(s.L, C.int(idx)) == C.LUA_TTHREAD
}

func (s *State) IsTable(idx int) bool {
	return C.lua_type(s.L, C.int(idx)) == C.LUA_TTABLE
}

func (s *State) IsUserData(idx int) bool {
	t := C.lua_type(s.L, C.int(idx))
	return t == C.LUA_TUSERDATA || t == C.LUA_TLIGHTUSERDATA
}

func (s *State) IsNil(idx int) bool {
	return C.lua_type(s.L, C.int(idx)) == C.LUA_TNIL
}

func (s *State) ObjLen(idx int) int {
	return int(C.lua_objlen(s.L, C.int(idx)))
}

func (s *State) ToBoolean(idx int) bool {
	return C.lua_toboolean(s.L, C.int(idx)) != 0
}

func (s *State) ToString(idx int) string {
	var len C.size_t
	cStr := C.lua_tolstring(s.L, C.int(idx), &len)
	return C.GoStringN(cStr, C.int(len))
}

func (s *State) ToNumber(idx int) float64 {
	return float64(C.lua_tonumber(s.L, C.int(idx)))
}

func (s *State) ToThread(idx int) *State {
	L := C.lua_tothread(s.L, C.int(idx))
	if L == nil {
		return nil
	}
	return &State{L: L}
}

// --- Table Operations ---

func (s *State) CreateTable(narr, nrec int) {
	C.lua_createtable(s.L, C.int(narr), C.int(nrec))
}

func (s *State) SetField(idx int, key string) {
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))
	C.lua_setfield(s.L, C.int(idx), cKey)
}

func (s *State) GetField(idx int, key string) {
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))
	C.lua_getfield(s.L, C.int(idx), cKey)
}

func (s *State) GetGlobal(key string) {
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))
	C.lua_getfield(s.L, C.LUA_GLOBALSINDEX, cKey)
}

func (s *State) SetGlobal(key string) {
	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))
	C.lua_setfield(s.L, C.LUA_GLOBALSINDEX, cKey)
}

func (s *State) RawSetInt(idx int, n int) {
	C.lua_rawseti(s.L, C.int(idx), C.int(n))
}

func (s *State) RawGetInt(idx int, n int) {
	C.lua_rawgeti(s.L, C.int(idx), C.int(n))
}

func (s *State) SetTable(idx int) {
	C.lua_settable(s.L, C.int(idx))
}

func (s *State) Next(idx int) bool {
	return C.lua_next(s.L, C.int(idx)) != 0
}

// --- Coroutine Operations ---

func (s *State) NewThread() *State {
	L := C.lua_newthread(s.L)
	return &State{L: L}
}

func (s *State) Resume(nargs int) (int, error) {
	status := C.lua_resume(s.L, C.int(nargs))
	if status != 0 && status != C.LUA_YIELD {
		return int(status), s.errorFromStack(status)
	}
	return int(status), nil
}

// Yield signals the C trampoline to yield.
// It returns the special code that C trampoline understands.
func (s *State) Yield(nresults int) int {
	// We map yield(n) to status code -(10 + n)
	// yield(0) -> -10 -> wait, my previous logic was:
	// -1: Yield 0
	// <= -10: Yield (status + 10) results?
	// Let's standardise on -1 for Yield(0).
	if nresults == 0 {
		return GO_YIELD_0
	}
	return -(10 + nresults)
}

// --- Registry & References ---

func (s *State) Ref(t int) int {
	return int(C.luaL_ref(s.L, C.int(t)))
}

func (s *State) Unref(t int, ref int) {
	C.luaL_unref(s.L, C.int(t), C.int(ref))
}

func (s *State) GetRef(ref int) {
	C.lua_rawgeti(s.L, C.LUA_REGISTRYINDEX, C.int(ref))
}

func (s *State) XMove(to *State, n int) {
	C.lua_xmove(s.L, to.L, C.int(n))
}

// Error signals the C trampoline to raise a Lua error.
func (s *State) Error() int {
	return GO_ERROR
}

// Pointer returns the underlying Lua state pointer
func (s *State) Pointer() uintptr {
	return uintptr(unsafe.Pointer(s.L))
}

// Flush flushes stdout and stderr
func (s *State) Flush() {
	C.fflush(C.stdout)
	C.fflush(C.stderr)
}

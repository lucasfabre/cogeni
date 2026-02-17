package luaruntime

import (
	"fmt"
	"sync"

	"github.com/lucasfabre/codegen/src/lua_runtime/luajit"
)

// Completion represents the result of an async operation
type Completion struct {
	Thread  *luajit.State
	Results []interface{}
}

// Scheduler manages async execution of Lua coroutines
type Scheduler struct {
	pending     chan Completion
	activeCount int // Number of active Go tasks

	anchors map[uintptr]int             // Registry references
	waiters map[uintptr][]*luajit.State // Target -> Waiters

	mu sync.Mutex
}

// NewScheduler creates a new scheduler
func NewScheduler() *Scheduler {
	return &Scheduler{
		pending: make(chan Completion, 100),
		anchors: make(map[uintptr]int),
		waiters: make(map[uintptr][]*luajit.State),
	}
}

// Register anchors a thread
func (s *Scheduler) Register(L *luajit.State) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ptr := L.Pointer()
	if _, ok := s.anchors[ptr]; ok {
		return
	}
	L.PushThread()
	ref := L.Ref(luajit.LUA_REGISTRYINDEX)
	s.anchors[ptr] = ref
}

// Unregister unanchors a thread
func (s *Scheduler) Unregister(L *luajit.State) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ptr := L.Pointer()
	if ref, ok := s.anchors[ptr]; ok {
		L.Unref(luajit.LUA_REGISTRYINDEX, ref)
		delete(s.anchors, ptr)
	}
}

// AsyncRun starts a Go function in a goroutine and yields the current Lua thread.
func (s *Scheduler) AsyncRun(L *luajit.State, task func() []interface{}) int {
	s.mu.Lock()
	s.activeCount++
	s.mu.Unlock()

	go func() {
		res := task()
		s.pending <- Completion{Thread: L, Results: res}
	}()

	return L.Yield(0)
}

// Await registers waiter to wait for target.
// Returns true if target is already dead (and results moved), false if waiter should yield.
func (s *Scheduler) Await(waiter *luajit.State, target *luajit.State) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	targetPtr := target.Pointer()
	if _, ok := s.anchors[targetPtr]; !ok {
		waiter.PushNil()
		return true
	}

	s.waiters[targetPtr] = append(s.waiters[targetPtr], waiter)
	return false
}

// Start runs the main Lua chunk
func (s *Scheduler) Start(mainL *luajit.State) error {
	// Local state for this run (fixes re-entrancy)
	mainFinished := false

	s.Register(mainL)

	status, err := s.resume(mainL, 0)
	if err != nil {
		return err
	}
	if status == luajit.OK {
		mainFinished = true
	}

	for {
		s.mu.Lock()
		activeGoTasks := s.activeCount
		anchoredThreads := len(s.anchors)
		s.mu.Unlock()

		// We finish when:
		// 1. Main thread is done (mainFinished=true)
		// 2. No active Go tasks (activeGoTasks=0)
		// 3. No anchored threads (anchoredThreads=0)
		if mainFinished && activeGoTasks == 0 && anchoredThreads == 0 {
			break
		}

		if activeGoTasks == 0 && anchoredThreads > 0 && !mainFinished {
			// Deadlock detection:
			// If main is not finished, it must be anchored (waiting).
			// If no Go tasks are active, nothing can wake it up (unless it's waiting on another thread).
			// If it's waiting on another thread, that thread must also be anchored.
			// If ALL threads are blocked and no Go tasks, it's a deadlock.
			// However, 'anchoredThreads' includes threads waiting on other threads.
			// So simple check: activeGoTasks == 0 && anchoredThreads > 0 is suspicious.
			// But we need to ensure ALL anchored threads are blocked.
			// Since Lua is single threaded, if we are here (Go runtime), Lua is not running.
			// So if no Go task is pending to resume a thread, we are stuck.
			return fmt.Errorf("deadlock detected: %d threads blocked with no active async tasks", anchoredThreads)
		} else if activeGoTasks == 0 && anchoredThreads > 0 && mainFinished {
			// Main finished but background threads still running (e.g. fire-and-forget async?)
			// If they are waiting for Go tasks, we should wait.
			// If activeGoTasks == 0, they are stuck.
			return fmt.Errorf("deadlock detected: main finished but %d threads blocked with no active async tasks", anchoredThreads)
		}

		select {
		case comp := <-s.pending:
			s.mu.Lock()
			s.activeCount--
			s.mu.Unlock()

			nResults := 0
			if comp.Results != nil {
				for _, res := range comp.Results {
					pushInterface(comp.Thread, res)
					nResults++
				}
			}

			st, err := s.resume(comp.Thread, nResults)
			if err != nil {
				return err
			}
			if comp.Thread == mainL && st == luajit.OK {
				mainFinished = true
			}
		}
	}
	return nil
}

func (s *Scheduler) resume(L *luajit.State, nArgs int) (int, error) {
	status, err := L.Resume(nArgs)
	if err != nil {
		s.Unregister(L)
		return 0, err
	}

	if status == luajit.OK {
		s.NotifyFinished(L)
		s.Unregister(L)
	}

	return status, nil
}

func (s *Scheduler) NotifyFinished(target *luajit.State) {
	s.mu.Lock()
	targetPtr := target.Pointer()
	waiters := s.waiters[targetPtr]
	delete(s.waiters, targetPtr)
	s.mu.Unlock()

	nResults := target.GetTop()

	for _, waiter := range waiters {
		for i := 1; i <= nResults; i++ {
			target.PushValue(i)
		}
		target.XMove(waiter, nResults)
		s.resume(waiter, nResults)
	}
}

func pushInterface(L *luajit.State, val interface{}) {
	PushGoValue(L, val)
}

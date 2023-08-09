package actor

import (
	"log/slog"
	"reflect"
	"sync"
)

type registry struct {
	mu     sync.RWMutex
	lookup map[string]Processor
	engine *Engine
}

func (r *registry) get(pid PID) Processor {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if proc, ok := r.lookup[pid.ID]; ok {
		return proc
	}

	return nil
}

func (r *registry) add(proc Processor) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := proc.PID().ID
	if existing, ok := r.lookup[id]; ok {
		slog.Warn("Attempt to register duplicate process.", "pid", proc.PID(), "existing", reflect.TypeOf(existing), "conflict", reflect.TypeOf(proc))
		return
	}

	r.lookup[id] = proc
}

func (r *registry) remove(pid PID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.lookup, pid.ID)
}

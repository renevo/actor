package actor

import "sync"

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
	if _, ok := r.lookup[id]; ok {
		// TODO: handle duplicates
		return
	}

	r.lookup[id] = proc
}

func (r *registry) remove(pid PID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.lookup, pid.ID)
}

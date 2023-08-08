package actor

import (
	"strings"
	"sync"
)

var (
	AddressSeparator = "."
)

const (
	LocalAddress = "local"
)

type Engine struct {
	pid        PID
	deadletter PID
	registry   *registry
}

func NewEngine() *Engine {
	e := &Engine{
		registry: &registry{
			lookup: make(map[string]Processor),
		},
	}

	// put the engine into the registry
	e.registry.engine = e
	e.pid = e.SpawnFunc(func(ctx *Context) {
		// TODO: engine stuff
	}, "engine")
	e.pid.Address = LocalAddress

	e.deadletter = e.SpawnFunc(func(ctx *Context) {
		// TODO: Deadletter stuff
	}, "engine", WithTags("deadletter"), WithInboxSize(defaultInboxSize*4))
	e.deadletter.Address = LocalAddress

	return e
}

func (e *Engine) Spawn(receiver Receiver, name string, opts ...Option) PID {
	options := DefaultOptions(receiver)
	options.ID = []string{name}
	for _, opt := range opts {
		opt(&options)
	}
	proc := newProcessor(e, options)
	return e.SpawnProcessor(proc)
}

func (e *Engine) SpawnFunc(receiver ReceiverFunc, name string, opts ...Option) PID {
	return e.Spawn(receiver, name, opts...)
}

func (e *Engine) SpawnProcessor(proc Processor) PID {
	e.registry.add(proc)
	proc.Start()
	return proc.PID()
}

func (e *Engine) Address() string {
	return e.pid.Address
}

func (e *Engine) SendWithSender(to PID, msg any, from PID) {
	e.send(to, msg, from)
}

func (e *Engine) Send(to PID, msg any) {
	e.send(to, msg, e.pid)
}

func (e *Engine) send(to PID, msg any, from PID) {
	proc := e.registry.get(to)
	if proc == nil {
		proc = e.registry.get(e.deadletter)
	}

	proc.Send(to, msg, from)
}

func (e *Engine) Poison(to PID, wg *sync.WaitGroup) {
	proc := e.registry.get(to)
	if proc == nil {
		return
	}

	if wg != nil {
		wg.Add(1)
	}

	e.send(to, poisonPill{wg: wg}, e.pid)
}

func (e *Engine) GetPID(id ...string) PID {
	pid := PID{Address: LocalAddress, ID: strings.Join(id, AddressSeparator)}
	proc := e.registry.get(pid)
	if proc == nil {
		return e.deadletter
	}

	return pid
}

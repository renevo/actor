package actor

import (
	"context"
	"log/slog"
	"reflect"
	"strings"
	"sync"
)

const (
	LocalAddress = "local"
	pidSeparator = "."
)

type Engine struct {
	pid        PID
	deadletter PID
	registry   *registry
	options    *Options
}

func NewEngine(defaultOpts ...Option) *Engine {
	options := &Options{
		InboxSize:    defaultInboxSize,
		MaxRestarts:  defaultMaxRestarts,
		RestartDelay: defaultRestartDelay,
		Context:      context.Background(),
		Logger:       slog.Default(),
	}
	for _, opt := range defaultOpts {
		opt(options)
	}

	e := &Engine{
		registry: &registry{
			lookup: make(map[string]Processor),
		},
		options: options,
	}

	// put the engine into the registry
	e.registry.engine = e
	e.pid = e.SpawnFunc(func(ctx *Context) {
		switch msg := ctx.Message().(type) {
		case Initialized, Started, Stopped:
			ctx.Log().Debug("engine state change", "state", reflect.TypeOf(msg))

		}
	}, "engine")

	e.deadletter = e.SpawnFunc(func(ctx *Context) {
		switch msg := ctx.Message().(type) {
		case Initialized, Started, Stopped:
			// if we have anything, add it here
			ctx.Log().Debug("deadletter state change", "state", reflect.TypeOf(msg))

		default:
			ctx.Log().Warn("Deadletter", "to", ctx.target, "from", ctx.sender, "type", reflect.TypeOf(ctx.Message()))
			// TODO: publish deadletter to Events once we have them
		}
	}, "engine", WithTags("deadletter"), WithInboxSize(defaultInboxSize*4))

	return e
}

func (e *Engine) Spawn(receiver Receiver, name string, opts ...Option) PID {
	options := copyOptions(e.options, receiver)
	options.Name = name
	for _, opt := range opts {
		opt(options)
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

func (e *Engine) Send(ctx context.Context, to PID, msg any) {
	e.send(ctx, to, msg, e.pid)
}

func (e *Engine) send(ctx context.Context, to PID, msg any, from PID) {
	proc := e.registry.get(to)
	if proc == nil {
		proc = e.registry.get(e.deadletter)
	}

	proc.Send(ctx, to, msg, from)
}

func (e *Engine) Poison(to PID, wg *sync.WaitGroup) {
	proc := e.registry.get(to)
	if proc == nil {
		return
	}

	if wg != nil {
		wg.Add(1)
	}

	// we intentionally use background here as we won't use it, so we dn't need it (yet)
	e.send(context.Background(), to, poisonPill{wg: wg}, e.pid)
}

func (e *Engine) GetPID(name string, tags ...string) PID {
	pid := PID{Address: LocalAddress, ID: strings.Join(append([]string{name}, tags...), pidSeparator)}
	proc := e.registry.get(pid)
	if proc == nil {
		return e.deadletter
	}

	return pid
}

func (e *Engine) Shutdown(wg *sync.WaitGroup) {
	var toShutdown []PID

	active := e.registry.copy()
	for pid, proc := range active {
		// we don't want the engine
		if pid.Equals(e.pid) {
			continue
		}

		// we don't want deadletter
		if pid.Equals(e.deadletter) {
			continue
		}

		internalProcessor, ok := proc.(*processor)
		if !ok {
			continue
		}

		// we only want actors that don't have a parent (i.e. root actors)
		if !internalProcessor.context.Parent().Equals(e.pid) {
			continue
		}

		toShutdown = append(toShutdown, pid)
	}

	var shutdownWG sync.WaitGroup
	for _, pid := range toShutdown {
		e.Poison(pid, &shutdownWG)
	}

	shutdownWG.Wait()

	// tell our engine/deadletter to die
	e.Poison(e.pid, wg)
	e.Poison(e.deadletter, wg)
}

func (e *Engine) ShutdownAndWait() {
	var wg sync.WaitGroup
	e.Shutdown(&wg)
	wg.Wait()
}

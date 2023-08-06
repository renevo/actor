package actor

import "context"

type Context struct {
	pid           PID
	sender        PID
	engine        *Engine
	receiver      Receiver
	message       any
	parentContext *Context
	ctx           context.Context
	children      *safemap[string, PID]
}

func newContext(e *Engine, pid PID) *Context {
	return &Context{
		engine:   e,
		pid:      pid,
		children: newMap[string, PID](),
	}
}

func (c *Context) Reciever() Receiver {
	return c.receiver
}

func (ctx *Context) Context() context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx.ctx
}

// Send message to another actor (core functionality)
func (ctx *Context) Send(to PID, msg any) {
	ctx.engine.SendWithSender(to, msg, ctx.pid)
}

// Forward the current message to another PID
func (ctx *Context) Forward(to PID) {
	ctx.engine.SendWithSender(to, ctx.message, ctx.pid)
}

func (ctx *Context) PID() PID {
	return ctx.pid
}

func (ctx *Context) Sender() PID {
	return ctx.sender
}

func (ctx *Context) Engine() *Engine {
	return ctx.engine
}

func (ctx *Context) Message() any {
	return ctx.message
}

func (ctx *Context) Parent() PID {
	if ctx.parentContext != nil {
		return ctx.parentContext.pid
	}

	return ctx.engine.pid
}

func (ctx *Context) Child(id string) (PID, bool) {
	return ctx.children.Get(id)
}

func (ctx *Context) Children() []PID {
	pids := make([]PID, ctx.children.Len())
	i := 0
	ctx.children.ForEach(func(_ string, child PID) {
		pids[i] = child
		i++
	})
	return pids
}

// Spawn a new actor (core functionality)
func (ctx *Context) Spawn(receiver Receiver, name string, opts ...Option) PID {
	options := DefaultOptions(receiver)
	options.ID = []string{ctx.PID().ID, name}
	for _, opt := range opts {
		opt(&options)
	}
	proc := newProcessor(ctx.engine, options)
	proc.context.parentContext = ctx
	pid := ctx.engine.SpawnProcessor(proc)
	ctx.children.Set(pid.ID, pid)

	return proc.PID()
}

// TODO: Add repeaters
// TODO: Add request/response

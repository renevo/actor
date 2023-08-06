package actor

import "context"

type Processor interface {
	Process(ctx context.Context, env Envelope)
}

type Receiver interface {
	Receive(ctx *Context)
}

type Context struct {
	ctx    context.Context
	engine *Engine
}

func (ctx *Context) Context() context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx.ctx
}

// Send message to another actor (core functionality)
func (ctx *Context) Send(addr Address, msg any) {
	ctx.engine.Deliver(Envelope{To: addr, Message: msg})
}

// Spawn a new actor (core functionality)
func (ctx *Context) Spawn() {

}

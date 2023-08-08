package actor

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

type Processor interface {
	PID() PID
	Start()
	Send(ctx context.Context, to PID, msg any, from PID)
	Process(env *Envelope)
	Shutdown(wg *sync.WaitGroup)
}

type processor struct {
	options  *Options
	inbox    *Inbox
	context  *Context
	pid      PID
	restarts int32
	pool     sync.Pool
}

func newProcessor(engine *Engine, opts *Options) *processor {
	pid := NewPID(engine.pid.Address, opts.Name, opts.Tags...)
	ctx := newContext(engine, pid)
	proc := &processor{
		pid:     pid,
		inbox:   NewInbox(opts.InboxSize),
		options: opts,
		context: ctx,
		// pool technically adds a tiny bit of CPU overhead (30ish ns), but will release some pressure on the GC
		pool: sync.Pool{
			New: func() any {
				return new(Envelope)
			},
		},
	}

	proc.inbox.Process(proc)

	return proc
}

func (p *processor) PID() PID {
	return p.pid
}

func (p *processor) Send(ctx context.Context, _ PID, msg any, from PID) {
	envelope := p.pool.Get().(*Envelope)
	envelope.To = p.pid
	envelope.From = from
	envelope.Message = msg
	envelope.Context = ctx
	if envelope.Context == nil {
		envelope.Context = p.context.engine.options.Context
	}

	if err := p.inbox.Deliver(envelope); err != nil {
		fmt.Fprintf(os.Stderr, "failed to deliver message to %s; from: %s; type: %T: %v\n", p.pid, from, msg, err)
	}
}

func (p *processor) Process(env *Envelope) {
	defer p.pool.Put(env)

	defer func() {
		if v := recover(); v != nil {
			// TODO: send this message to a poison processor with the error associated with it
			p.context.ctx = p.context.engine.options.Context
			p.context.message = Stopped{}
			p.context.receiver.Receive(p.context)

			if p.options.MaxRestarts > 0 {
				p.tryRestart(v)
			}
		}
	}()

	if pill, ok := env.Message.(poisonPill); ok {
		p.cleanup(pill.wg)
		return
	}

	p.context.message = env.Message
	p.context.sender = env.From
	p.context.ctx = env.Context

	// TODO: Check to see if context.Deadline and drop if Options say we should
	rcv := p.context.receiver

	p.applyMiddleware(rcv.Receive, p.options.Middleware...)(p.context)
}

func (p *processor) Start() {
	rcv := p.options.Receiver
	p.context.receiver = rcv

	p.context.ctx = p.context.engine.options.Context
	p.context.message = Initialized{}
	p.applyMiddleware(rcv.Receive, p.options.Middleware...)(p.context)

	p.context.ctx = p.context.engine.options.Context
	p.context.message = Started{}
	p.applyMiddleware(rcv.Receive, p.options.Middleware...)(p.context)
}

func (p *processor) Shutdown(wg *sync.WaitGroup) {
	p.cleanup(wg)
}

func (p *processor) applyMiddleware(rcv ReceiverFunc, middleware ...Middleware) ReceiverFunc {
	for i := len(middleware) - 1; i >= 0; i-- {
		rcv = middleware[i](rcv)
	}
	return rcv
}

func (p *processor) cleanup(wg *sync.WaitGroup) {
	p.inbox.Close()
	p.context.engine.registry.remove(p.pid)

	p.context.ctx = p.context.engine.options.Context
	p.context.message = Stopped{}
	p.applyMiddleware(p.context.receiver.Receive, p.options.Middleware...)(p.context)

	if p.context.parentContext != nil {
		p.context.parentContext.children.Delete(p.pid.ID)
	}

	if p.context.children.Len() > 0 {
		children := p.context.Children()
		for _, pid := range children {
			proc := p.context.engine.registry.get(pid)
			if proc == nil {
				continue
			}

			if wg != nil {
				wg.Add(1)
			}

			proc.Shutdown(wg)
		}
	}

	// send events
	if wg != nil {
		wg.Done()
	}
}

func (p *processor) tryRestart(v any) {
	p.restarts++

	if p.restarts >= p.options.MaxRestarts {
		fmt.Fprintf(os.Stderr, "Process max restarts exceeded, shutting down: pid: %s; restarts: %d\n", p.pid, p.restarts)
		p.cleanup(nil)
		return
	}

	fmt.Fprintf(os.Stderr, "Process actor restarting: count: %d; maxRestarts: %d; pid: %s; reason: %v\n", p.restarts, p.options.MaxRestarts, p.pid, v)
	time.Sleep(p.options.RestartDelay)
	p.Start()
}

package actor

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type PID struct {
	Address string
	ID      string
}

func NewPID(address string, id ...string) PID {
	return PID{
		Address: address,
		ID:      strings.Join(id, AddressSeparator),
	}
}

func (p PID) Equals(pid PID) bool {
	return pid.Address == p.Address && pid.ID == p.ID
}

func (p PID) String() string {
	return p.Address + AddressSeparator + p.ID
}

func (p PID) Child(id ...string) PID {
	return NewPID(p.Address, p.ID+AddressSeparator+strings.Join(id, AddressSeparator))
}

func (p PID) IsZero() bool {
	return p.Address == "" && p.ID == ""
}

type Processor interface {
	PID() PID
	Start()
	Send(to PID, msg any, from PID)
	Process(ctx context.Context, env Envelope)
	Shutdown(wg *sync.WaitGroup)
}

type processor struct {
	Options
	inbox    *Inbox
	context  *Context
	pid      PID
	restarts int32
}

func newProcessor(engine *Engine, opts Options) *processor {
	pid := NewPID(engine.pid.Address, opts.ID...)
	ctx := newContext(engine, pid)
	proc := &processor{
		pid:     pid,
		inbox:   NewInbox(opts.InboxSize),
		Options: opts,
		context: ctx,
	}

	proc.inbox.Process(opts.BaseContext, proc)

	return proc
}

func (p *processor) PID() PID {
	return p.pid
}

func (p *processor) Send(_ PID, msg any, from PID) {
	if err := p.inbox.Deliver(Envelope{To: p.pid, Message: msg, From: from}); err != nil {
		fmt.Fprintf(os.Stderr, "failed to deliver message to %s; from: %s; type: %T: %v\n", p.pid, from, msg, err)
	}
}

func (p *processor) Process(ctx context.Context, env Envelope) {
	defer func() {
		if v := recover(); v != nil {
			p.context.message = Stopped{}
			p.context.receiver.Receive(p.context)

			if p.Options.MaxRestarts > 0 {
				p.tryRestart(v)
			}
		}
	}()

	if pill, ok := env.Message.(poisonPill); ok {
		p.cleanup(pill.wg)
		return
	}

	p.context.ctx = ctx
	p.context.message = env.Message
	p.context.sender = env.From
	rcv := p.context.receiver

	p.applyMiddleware(rcv.Receive, p.Options.Middleware...)(p.context)
}

func (p *processor) Start() {
	rcv := p.Receiver
	p.context.receiver = rcv

	p.context.message = Initialized{}
	p.applyMiddleware(rcv.Receive, p.Options.Middleware...)(p.context)

	p.context.message = Started{}
	p.applyMiddleware(rcv.Receive, p.Options.Middleware...)(p.context)
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

	p.context.message = Stopped{}
	p.applyMiddleware(p.context.receiver.Receive, p.Options.Middleware...)(p.context)

	if p.context.parentContext != nil {
		p.context.parentContext.children.Delete(p.pid.ID)
	}

	if p.context.children.Len() > 0 {
		children := p.context.Children()
		for _, pid := range children {
			if wg != nil {
				wg.Add(1)
			}
			proc := p.context.engine.registry.get(pid)

			// don't tell the deadletter to shutdown...
			if proc.PID() == p.context.engine.deadletter {
				continue
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

	if p.restarts >= p.Options.MaxRestarts {
		fmt.Fprintf(os.Stderr, "Process max restarts exceeded, shutting down: pid: %s; restarts: %d\n", p.pid, p.restarts)
		p.cleanup(nil)
		return
	}

	fmt.Fprintf(os.Stderr, "Process actor restarting: count: %d; maxRestarts: %d; pid: %s; reason: %v\n", p.restarts, p.Options.MaxRestarts, p.pid, v)
	time.Sleep(p.Options.RestartDelay)
	p.Start()
}

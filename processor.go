package actor

import (
	"context"
	"reflect"
	"sync"
	"time"
)

type processorState byte

var (
	processorStateCreated     processorState = 1
	processorStateInitialized processorState = 2
	processorStateStarted     processorState = 3
	processorStateStopped     processorState = 4
)

var (
	envelopePool = sync.Pool{
		New: func() any {
			return new(Envelope)
		},
	}
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
	restarts int
	state    processorState
	init     sync.Once
}

func newProcessor(engine *Engine, opts *Options) *processor {
	pid := NewPID(engine.pid.Address, opts.Name, opts.Tags...)
	proc := &processor{
		state:   processorStateCreated,
		pid:     pid,
		inbox:   NewInbox(opts.InboxSize),
		options: opts,
		context: newContext(engine, pid),
	}

	return proc
}

func (p *processor) PID() PID {
	return p.pid
}

func (p *processor) Send(ctx context.Context, to PID, msg any, from PID) {
	env := envelopePool.Get().(*Envelope)
	env.To = to
	env.From = from
	env.Message = msg
	env.Context = ctx
	if env.Context == nil {
		env.Context = p.context.engine.options.Context
	}

	if err := p.inbox.Deliver(env); err != nil {
		p.context.logger.Error("Failed to deliver message to inbox.", "inbox", p.pid, "from", from, "msg", reflect.TypeOf(msg), "err", err)
	}
}

func (p *processor) Process(env *Envelope) {
	defer envelopePool.Put(env)

	// uncomment this for low level message process logging, its super spammy if left on, even at trace level
	// p.context.logger.Info("processor.Process", "backlog", len(p.inbox.box), "to", env.To, "from", env.From, "msg", reflect.TypeOf(env.Message), "content", env.Message)

	defer func() {
		if v := recover(); v != nil {
			// TODO: send this message to a poison processor with the error associated with it

			// only send the stop if in a valid state to actually stop
			if p.state == processorStateStarted {
				p.context.ctx = p.context.engine.options.Context
				p.context.message = Stopped{}
				p.context.receiver.Receive(p.context)
			}

			p.state = processorStateStopped

			if p.options.MaxRestarts > 0 {
				p.tryRestart(v)
			}
		}
	}()

	// TODO: Check to see if context.Deadline and drop if Options say we should
	rcv := p.context.receiver

	if p.state == processorStateCreated || p.state == processorStateStopped {
		p.context.ctx = p.context.engine.options.Context
		p.context.message = Initialized{}
		p.applyMiddleware(rcv.Receive, p.options.Middleware...)(p.context)
		p.state = processorStateInitialized
	}

	if p.state == processorStateInitialized {
		p.context.ctx = p.context.engine.options.Context
		p.context.message = Started{}
		p.applyMiddleware(rcv.Receive, p.options.Middleware...)(p.context)
		p.state = processorStateStarted
	}

	// just kicking off
	if _, ok := env.Message.(initialize); ok {
		return
	}

	if pill, ok := env.Message.(poisonPill); ok {
		p.cleanup(pill.wg)
		return
	}

	p.context.message = env.Message
	p.context.sender = env.From
	p.context.target = env.To
	p.context.ctx = env.Context

	p.applyMiddleware(rcv.Receive, p.options.Middleware...)(p.context)
}

func (p *processor) Start() {
	p.context.receiver = p.options.Receiver

	// kick off initialize of the actor if it hasn't already
	p.init.Do(func() {
		p.inbox.Process(p)
		p.Send(context.Background(), p.pid, initialize{}, p.pid)
	})
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
	p.context.engine.registry.remove(p.pid)
	p.inbox.Close()

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

	// only send the stop if in a valid state to actually stop
	if p.state == processorStateStarted {
		p.context.ctx = p.context.engine.options.Context
		p.context.message = Stopped{}
		p.applyMiddleware(p.context.receiver.Receive, p.options.Middleware...)(p.context)
	}

	p.state = processorStateStopped

	// send events
	if wg != nil {
		wg.Done()
	}
}

func (p *processor) tryRestart(v any) {
	p.restarts++

	if p.restarts > p.options.MaxRestarts {
		p.context.logger.Error("Actor process max restarts exceeded, shutting down.", "restarts", p.restarts, "err", v)
		p.cleanup(nil)
		return
	}

	p.context.logger.Warn("Actor process restarting.", "restarts", p.restarts, "maxRestarts", p.options.MaxRestarts, "err", v)
	time.Sleep(p.options.RestartDelay)

	if len(p.inbox.box) == 0 {
		p.Send(context.Background(), p.pid, initialize{}, p.pid)
	}
}

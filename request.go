package actor

import (
	"context"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

type response struct {
	engine  *Engine
	ctx     context.Context
	pid     PID
	result  chan any
	timeout time.Duration
}

func newResponse(ctx context.Context, engine *Engine, timeout time.Duration) *response {
	r := &response{
		engine:  engine,
		ctx:     ctx,
		result:  make(chan any, 1),
		timeout: timeout,
		pid:     NewPID(engine.pid.Address, "response", strconv.Itoa(rand.Intn(100_000))),
	}

	if r.ctx == nil {
		r.ctx = context.Background()
	}

	return r
}

func (r *response) waitForResult() (any, error) {
	ctx, cancel := context.WithTimeout(r.ctx, r.timeout)
	defer func() {
		cancel()
		r.engine.registry.remove(r.pid)
	}()

	select {
	case resp := <-r.result:
		return resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (r *response) Send(ctx context.Context, _ PID, msg any, _ PID) {
	r.result <- msg
}

func (r *response) PID() PID                   { return r.pid }
func (r *response) Shutdown(_ *sync.WaitGroup) {}
func (r *response) Start()                     {}
func (r *response) Process(*Envelope)          {}

func (e *Engine) Request(to PID, msg any, timeout time.Duration) (any, error) {
	resp := newResponse(e.options.Context, e, timeout)
	e.registry.add(resp)
	e.send(e.options.Context, to, msg, resp.PID())
	return resp.waitForResult()
}

func (c *Context) Request(to PID, msg any, timeout time.Duration) (any, error) {
	resp := newResponse(c.ctx, c.engine, timeout)
	c.engine.registry.add(resp)
	c.engine.send(c.ctx, to, msg, resp.PID())
	return resp.waitForResult()
}

func (c *Context) Respond(msg any) {
	if c.sender.IsZero() {
		// TODO: Log about responding to no one
		return
	}

	c.Send(c.ctx, c.sender, msg)
}

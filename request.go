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
	pid     PID
	result  chan any
	timeout time.Duration
}

func newResponse(engine *Engine, timeout time.Duration) *response {
	return &response{
		engine:  engine,
		result:  make(chan any, 1),
		timeout: timeout,
		pid:     NewPID(engine.pid.Address, "response", strconv.Itoa(rand.Intn(100_000))),
	}
}

func (r *response) waitForResult() (any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
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

func (r *response) Send(_ PID, msg any, _ PID) {
	r.result <- msg
}

func (r *response) PID() PID                          { return r.pid }
func (r *response) Shutdown(_ *sync.WaitGroup)        {}
func (r *response) Start()                            {}
func (r *response) Process(context.Context, Envelope) {}

func (e *Engine) Request(to PID, msg any, timeout time.Duration) (any, error) {
	resp := newResponse(e, timeout)
	e.registry.add(resp)
	e.SendWithSender(to, msg, resp.PID())
	return resp.waitForResult()
}

func (c *Context) Request(to PID, msg any, timeout time.Duration) (any, error) {
	return c.engine.Request(to, msg, timeout)
}

func (c *Context) Respond(msg any) {
	if c.sender.IsZero() {
		// TODO: Log about responding to no one
		return
	}

	c.Send(c.sender, msg)
}

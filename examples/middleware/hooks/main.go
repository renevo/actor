package main

import (
	"context"

	"github.com/renevo/actor"
)

type Hooker interface {
	OnInit(*actor.Context)
	OnStart(*actor.Context)
	OnStop(*actor.Context)
	Receive(*actor.Context)
}

type foo struct{}

func newFoo() actor.Receiver {
	return &foo{}
}

func (f *foo) Receive(ctx *actor.Context) {}
func (f *foo) OnInit(ctx *actor.Context)  { ctx.Log().Info("foo initialized") }
func (f *foo) OnStart(ctx *actor.Context) { ctx.Log().Info("foo started") }
func (f *foo) OnStop(ctx *actor.Context)  { ctx.Log().Info("foo stopped") }

func WithHooks() func(actor.ReceiverFunc) actor.ReceiverFunc {
	return func(next actor.ReceiverFunc) actor.ReceiverFunc {
		return func(c *actor.Context) {
			switch c.Message().(type) {
			case actor.Initialized:
				c.Receiver().(Hooker).OnInit(c)
			case actor.Started:
				c.Receiver().(Hooker).OnStart(c)
			case actor.Stopped:
				c.Receiver().(Hooker).OnStop(c)
			}
			next(c)
		}
	}
}

func main() {
	// Create a new engine
	e := actor.NewEngine()

	// Spawn the a new "foo" receiver with middleware.
	pid := e.Spawn(newFoo(), "foo", actor.WithMiddleware(WithHooks()))

	// Send a message to foo
	e.Send(context.Background(), pid, "Hello sailor!")

	// calling ShutdownAndWait will guarantee that all messages before this call are processed
	e.ShutdownAndWait()
}

package main

import (
	"context"

	"github.com/renevo/actor"
)

type message struct {
	data string
}

type foo struct{}

func newFoo() actor.Receiver {
	return &foo{}
}

func (f *foo) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case actor.Initialized:
		ctx.Log().Info("foot has initialized")
	case actor.Started:
		ctx.Log().Info("foo has started")
	case *message:
		ctx.Log().Info("foo has received", "data", msg.data)
	case actor.Stopped:
		ctx.Log().Info("foo has stopped")
	}
}

func main() {
	engine := actor.NewEngine()
	pid := engine.Spawn(newFoo(), "foo")
	for i := 0; i < 10; i++ {
		engine.Send(context.Background(), pid, &message{data: "hello world!"})
	}

	engine.ShutdownAndWait()
}

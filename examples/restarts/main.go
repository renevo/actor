package main

import (
	"context"
	"time"

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
	case actor.Started:
		ctx.Log().Info("foo started")
	case *message:
		if msg.data == "failed" {
			panic("I failed processing this message")
		}
		ctx.Log().Info("I restarted and processed the next one perfectly", "data", msg.data)
	}
}

func main() {
	engine := actor.NewEngine(actor.WithRestartDelay(time.Second * 1))
	pid := engine.Spawn(newFoo(), "foo", actor.WithMaxRestarts(3))
	engine.Send(context.Background(), pid, &message{data: "failed"})
	engine.Send(context.Background(), pid, &message{data: "failed"})
	engine.Send(context.Background(), pid, &message{data: "failed"})
	engine.Send(context.Background(), pid, &message{data: "hello world!"})

	engine.ShutdownAndWait()
}

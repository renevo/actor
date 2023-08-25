package main

import (
	"context"
	"fmt"

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
		fmt.Println("foot has initialized")
	case actor.Started:
		fmt.Println("foo has started")
	case *message:
		fmt.Println("foo has received", msg.data)
	case actor.Stopped:
		fmt.Println("foo has stopped")
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

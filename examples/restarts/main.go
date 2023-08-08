package main

import (
	"fmt"
	"sync"

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
		fmt.Println("foo started")
	case *message:
		if msg.data == "failed" {
			panic("I failed processing this message")
		}
		fmt.Println("I restarted and processed the next one perfectly:", msg.data)
	}
}

func main() {
	engine := actor.NewEngine()
	pid := engine.Spawn(newFoo(), "foo", actor.WithMaxRestarts(3))
	engine.Send(pid, &message{data: "failed"})
	engine.Send(pid, &message{data: "hello world!"})

	wg := &sync.WaitGroup{}
	engine.Poison(pid, wg)
	wg.Wait()
}

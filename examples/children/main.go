package main

import (
	"context"
	"fmt"

	"github.com/renevo/actor"
)

var restarts = 0

type barReceiver struct {
	data string
}

func newBarReceiver(data string) actor.Receiver {
	return &barReceiver{
		data: data,
	}
}

func (r *barReceiver) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case actor.Started:
		ctx.Log().Info("started")
		if restarts < 2 {
			restarts++
			panic("I will need to restart")
		}
		ctx.Log().Warn("bar recovered and started with initial state", "data", r.data)
	case message:
		ctx.Log().Info(msg.data)
	case actor.Stopped:
		ctx.Log().Info("stopped")
	}
}

type fooReceiver struct {
	barPID actor.PID
}

func newFooReceiver() actor.Receiver {
	return &fooReceiver{}
}

func (r *fooReceiver) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case actor.Started:
		ctx.Log().Info("started")
	case message:
		if r.barPID.ID == "" {
			// this is a fundamental flaw in the system
			// if bar panics, it bubbles the panic up to this actor instead of the child
			// this then also creates an entry in the registry that has a borked up processor, and we don't get the pid back here
			// might need to make the inbox a pull rather than a push, and processor.Start kick off a go routine instead of the go routine being in the inbox
			r.barPID = ctx.Spawn(newBarReceiver(msg.data), "bar", actor.WithTags(msg.data))
			ctx.Log().Info("received and starting bar", "bar", r.barPID)
		}
	case actor.Stopped:
		ctx.Log().Info("will stop")
	}
}

type message struct {
	data string
}

func main() {
	e := actor.NewEngine()
	pid := e.Spawn(newFooReceiver(), "foo")
	e.Send(context.Background(), pid, message{data: fmt.Sprintf("msg_%d", 1)})
	e.Send(context.Background(), pid, message{data: fmt.Sprintf("msg_%d", 1)})
	e.Send(context.Background(), pid, message{data: fmt.Sprintf("msg_%d", 1)})

	e.ShutdownAndWait()
}

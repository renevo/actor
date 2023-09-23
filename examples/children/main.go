package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/renevo/actor"
)

var (
	restarts = 0
	wg       sync.WaitGroup
)

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
	case actor.Initialized:
		ctx.Log().Info("initialized")
	case actor.Started:
		ctx.Log().Info("started")
		if restarts < 2 {
			restarts++
			panic("I will need to restart")
		}
		ctx.Log().Warn("bar recovered and started with initial state", "data", r.data)
		wg.Done()

	case string:
		r.data = msg
		ctx.Log().Info("Message", "data", msg)
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
			r.barPID = ctx.Spawn(newBarReceiver(msg.data), "bar")
			ctx.Log().Info("received and starting bar", "bar", r.barPID)
		}

		ctx.Send(ctx.Context(), r.barPID, msg.data)

	case actor.Stopped:
		ctx.Log().Info("stopped")
	}
}

type message struct {
	data string
}

func main() {
	wg.Add(1)
	e := actor.NewEngine()
	pid := e.Spawn(newFooReceiver(), "foo")
	e.Send(context.Background(), pid, message{data: fmt.Sprintf("msg_%d", 1)})
	e.Send(context.Background(), pid, message{data: fmt.Sprintf("msg_%d", 2)})
	e.Send(context.Background(), pid, message{data: fmt.Sprintf("msg_%d", 3)})
	wg.Wait()

	e.ShutdownAndWait()
}

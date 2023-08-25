package main

import (
	"context"

	"github.com/renevo/actor"
)

func main() {
	engine := actor.NewEngine()

	// TODO: need to show how to capture the event, as well as explain this a bit better...
	engine.Send(context.Background(), actor.NewPID(actor.LocalAddress, "actor", "missing"), "hello")

	engine.ShutdownAndWait()
}

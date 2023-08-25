package main

import (
	"log"
	"log/slog"
	"time"

	"github.com/renevo/actor"
)

type (
	nameRequest struct{}

	nameResponse struct {
		name string
	}
)

type nameResponder struct{}

func newNameResponder() actor.Receiver {
	return &nameResponder{}
}

func (r *nameResponder) Receive(ctx *actor.Context) {
	switch ctx.Message().(type) {
	case *nameRequest:
		ctx.Respond(&nameResponse{name: "renevo"})
	}
}

func main() {
	engine := actor.NewEngine()
	pid := engine.Spawn(newNameResponder(), "responder")
	// Request a name and block till we got a response or the request timed out.
	res, err := engine.Request(pid, &nameRequest{}, time.Millisecond)
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("Response", "data", res)
}

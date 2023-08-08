package main

import (
	"fmt"
	"log"
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
	e := actor.NewEngine()
	pid := e.Spawn(newNameResponder(), "responder")
	// Request a name and block till we got a response or the request timed out.
	res, err := e.Request(pid, &nameRequest{}, time.Millisecond)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(res)
}

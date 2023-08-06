package actor_test

import (
	"context"
	"sync"

	"github.com/renevo/actor"
)

type testProcessor struct {
	processFn func(ctx context.Context, env actor.Envelope)
}

func (tp *testProcessor) Process(ctx context.Context, env actor.Envelope) {
	tp.processFn(ctx, env)
}

func (tp *testProcessor) PID() actor.PID {
	return actor.NewPID(actor.LocalAddress, "test")
}

func (tp *testProcessor) Start() {

}

func (tp *testProcessor) Send(to actor.PID, msg any, from actor.PID) {

}

func (tp *testProcessor) Shutdown(wg *sync.WaitGroup) {

}

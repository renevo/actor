package actor_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/matryer/is"
	"github.com/renevo/actor"
)

type testProcessor struct {
	processFn func(env *actor.Envelope)
}

func (tp *testProcessor) Process(env *actor.Envelope) {
	tp.processFn(env)
}

func (tp *testProcessor) PID() actor.PID {
	return actor.NewPID(actor.LocalAddress, "test")
}

func (tp *testProcessor) Start() {

}

func (tp *testProcessor) Send(ctx context.Context, to actor.PID, msg any, from actor.PID) {

}

func (tp *testProcessor) Shutdown(wg *sync.WaitGroup) {

}

func TestInbox(t *testing.T) {
	is := is.New(t)

	inbox := actor.NewInbox(128)
	inbox.Process(&testProcessor{processFn: func(env *actor.Envelope) {
		t.Logf("TO: %+v; From: %+v; Message: %+v;", env.To, env.From, env.Message)
	}})

	to := actor.NewPID(actor.LocalAddress, "to")
	from := actor.NewPID(actor.LocalAddress, "from")

	for i := 1; i <= 100; i++ {
		is.NoErr(inbox.Deliver(&actor.Envelope{To: to, From: from, Message: fmt.Sprintf("Hello: %d", i)}))
	}

	// closes and waits for drain
	inbox.Drain()

	// this should not panic
	inbox.Close()

	// this should return closed inbox
	for i := 1; i <= 100; i++ {
		is.Equal(actor.ErrInboxClosed, inbox.Deliver(&actor.Envelope{To: to, From: from, Message: fmt.Sprintf("Hello: %d", i)}))
	}
}

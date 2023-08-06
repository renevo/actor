package actor_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/renevo/actor"
)

type testProcessor struct {
	processFn func(ctx context.Context, env actor.Envelope)
}

func (tp *testProcessor) Process(ctx context.Context, env actor.Envelope) {
	tp.processFn(ctx, env)
}

func TestInbox(t *testing.T) {
	inbox := actor.NewInbox(128)
	inbox.Process(context.Background(), &testProcessor{processFn: func(ctx context.Context, env actor.Envelope) {
		t.Logf("TO: %+v; From: %+v; Message: %+v;", env.To, env.From, env.Message)
	}})

	for i := 1; i <= 100; i++ {
		if err := inbox.Deliver(actor.Envelope{To: actor.Address{"to"}, From: actor.Address{"from"}, Message: fmt.Sprintf("Hello: %d", i)}); err != nil {
			t.Errorf("Deliver Failure: %v", err)
		}
	}

	inbox.Close()

	// this should not panic
	inbox.Close()

	// this should return closed inbox
	for i := 1; i <= 100; i++ {
		if err := inbox.Deliver(actor.Envelope{To: actor.Address{"to"}, From: actor.Address{"from"}, Message: fmt.Sprintf("Hello: %d", i)}); !errors.Is(err, actor.ErrInboxClosed) {
			t.Errorf("Unexpected Deliver Error: %v", err)
		}
	}
}

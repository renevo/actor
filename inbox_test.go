package actor_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/renevo/actor"
)

func TestInbox(t *testing.T) {
	inbox := actor.NewInbox(128)
	inbox.Process(context.Background(), &testProcessor{processFn: func(ctx context.Context, env actor.Envelope) {
		t.Logf("TO: %+v; From: %+v; Message: %+v;", env.To, env.From, env.Message)
	}})

	to := actor.NewPID(actor.LocalAddress, "to")
	from := actor.NewPID(actor.LocalAddress, "from")

	for i := 1; i <= 100; i++ {
		if err := inbox.Deliver(actor.Envelope{To: to, From: from, Message: fmt.Sprintf("Hello: %d", i)}); err != nil {
			t.Errorf("Deliver Failure: %v", err)
		}
	}

	// closes and waits for drain
	inbox.Drain()

	// this should not panic
	inbox.Close()

	// this should return closed inbox
	for i := 1; i <= 100; i++ {
		if err := inbox.Deliver(actor.Envelope{To: to, From: from, Message: fmt.Sprintf("Hello: %d", i)}); !errors.Is(err, actor.ErrInboxClosed) {
			t.Errorf("Unexpected Deliver Error: %v", err)
		}
	}
}

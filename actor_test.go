package actor_test

import (
	"testing"

	"github.com/renevo/actor"
)

type TestReceiveFunc func(*testing.T, *actor.Context)

type TestReceiver struct {
	OnReceive TestReceiveFunc
	t         *testing.T
}

func NewTestReceiver(t *testing.T, f TestReceiveFunc) actor.Receiver {
	return &TestReceiver{
		OnReceive: f,
		t:         t,
	}
}

func (r *TestReceiver) Receive(ctx *actor.Context) {
	r.OnReceive(r.t, ctx)
}

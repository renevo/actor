package actor_test

import (
	"context"
	"sync"
	"testing"

	"github.com/matryer/is"
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

func TestMiddleware(t *testing.T) {
	is := is.New(t)

	engine := actor.NewEngine()
	callCount := 0

	pid := engine.SpawnFunc(func(ctx *actor.Context) {
		switch ctx.Message().(type) {
		case string:
			t.Logf("Hello, %s! (%T)", ctx.Message(), ctx.Message())
		}
	}, "middleware-tester",
		actor.WithMiddleware(
			func(next actor.ReceiverFunc) actor.ReceiverFunc {
				return func(ctx *actor.Context) {
					callCount++
					next(ctx)
				}
			},
		))

	engine.Send(context.Background(), pid, "hello")

	wg := &sync.WaitGroup{}
	engine.Poison(pid, wg)
	wg.Wait()

	is.Equal(callCount, 4) // middleware was not called correct amount of times
}

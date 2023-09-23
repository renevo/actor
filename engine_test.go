package actor_test

import (
	"context"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/renevo/actor"
)

func TestRestarts(t *testing.T) {
	engine := actor.NewEngine()
	type payload struct {
		data int
	}

	var panicInInit bool
	var panicInStart bool

	pid := engine.SpawnFunc(func(ctx *actor.Context) {
		switch msg := ctx.Message().(type) {
		case actor.Initialized:
			if !panicInInit {
				panicInInit = true
				panic("panic in init")
			}

			ctx.Log().Info("initd!")
		case actor.Started:
			if !panicInStart {
				panicInStart = true
				panic("panic in start")
			}

			ctx.Log().Info("started!")

		case actor.Stopped:
			ctx.Log().Info("stopped!")

		case payload:
			if msg.data != 10 {
				panic("I failed to process this message")
			} else {
				ctx.Log().Info("finally processed all my messsages after borking", "data", msg.data)
			}
		}
	}, "TestRestarts", actor.WithRestartDelay(time.Millisecond*10), actor.WithTags("bar"), actor.WithMaxRestarts(5))

	engine.Send(context.Background(), pid, payload{1})
	engine.Send(context.Background(), pid, payload{2})
	engine.Send(context.Background(), pid, payload{10})
	engine.ShutdownAndWait()
}

func TestProcessInitStartOrder(t *testing.T) {
	is := is.New(t)

	engine := actor.NewEngine()
	wg := &sync.WaitGroup{}
	var started, init bool
	wg.Add(1)

	pid := engine.SpawnFunc(func(ctx *actor.Context) {
		switch ctx.Message().(type) {
		case actor.Initialized:
			ctx.Log().Info("init", "init", init, "started", started)
			init = true
		case actor.Started:
			ctx.Log().Info("start", "init", init, "started", started)
			is.True(init)
			started = true
		case int:
			ctx.Log().Info("msg", "init", init, "started", started)
			is.True(started)
			wg.Done()
		}
	}, "TestProcessInitStartOrder")

	engine.Send(context.Background(), pid, 1)
	wg.Wait()
	is.True(init)
	is.True(started)
}

func TestSendMsgRaceCon(t *testing.T) {
	engine := actor.NewEngine()
	wg := &sync.WaitGroup{}

	pid := engine.SpawnFunc(func(ctx *actor.Context) {
		msg := ctx.Message()
		if msg == nil {
			ctx.Log().Error("should never happen")
		}
	}, "TestSendMsgRaceCon")

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			engine.Send(context.Background(), pid, []byte("f"))
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestSpawn(t *testing.T) {
	engine := actor.NewEngine()
	wg := &sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			tag := strconv.Itoa(i)
			pid := engine.Spawn(NewTestReceiver(t, func(t *testing.T, ctx *actor.Context) {
			}), "TestSpawn", actor.WithTags(tag))
			engine.Send(context.Background(), pid, 1)
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func TestPoisonWaitGroup(t *testing.T) {
	is := is.New(t)

	engine := actor.NewEngine()
	wg := sync.WaitGroup{}
	x := int32(0)

	wg.Add(1)
	pid := engine.SpawnFunc(func(ctx *actor.Context) {
		switch ctx.Message().(type) {
		case actor.Started:
			wg.Done()
		case actor.Stopped:
			atomic.AddInt32(&x, 1)
		}
	}, "TestPoisonWaitGroup")
	wg.Wait()

	pwg := &sync.WaitGroup{}
	engine.Poison(pid, pwg)
	pwg.Wait()
	is.Equal(int32(1), atomic.LoadInt32(&x))
}

type tick struct{}
type tickReceiver struct {
	ticks int
	wg    *sync.WaitGroup
}

func (r *tickReceiver) Receive(c *actor.Context) {
	switch c.Message().(type) {
	case tick:
		r.ticks++
		if r.ticks == 10 {
			r.wg.Done()
		}
	}
}

func newTickReceiver(wg *sync.WaitGroup) actor.Receiver {
	return &tickReceiver{
		wg: wg,
	}
}

func TestSendRepeat(t *testing.T) {
	engine := actor.NewEngine()
	wg := &sync.WaitGroup{}

	wg.Add(1)
	pid := engine.Spawn(newTickReceiver(wg), "TestSendRepeat")
	repeater := engine.SendRepeat(pid, tick{}, time.Millisecond*2)
	wg.Wait()
	repeater.Stop()
}

func TestRequestResponse(t *testing.T) {
	is := is.New(t)

	engine := actor.NewEngine()
	pid := engine.Spawn(NewTestReceiver(t, func(t *testing.T, ctx *actor.Context) {
		if msg, ok := ctx.Message().(string); ok {
			is.Equal("foo", msg)
			ctx.Respond("bar")
		}
	}), "TestRequestResponse")

	resp, err := engine.Request(pid, "foo", time.Millisecond)

	is.NoErr(err)
	is.Equal("bar", resp)
}

func BenchmarkEngineSend(b *testing.B) {
	engine := actor.NewEngine()
	pid := engine.SpawnFunc(func(*actor.Context) {}, "BenchmarkEngineSend", actor.WithInboxSize(1024*8))
	ctx := context.Background()
	payload := make([]byte, 128)
	wg := &sync.WaitGroup{}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		engine.Send(ctx, pid, payload)
	}

	engine.Poison(pid, wg)
	wg.Wait()
}

func BenchmarkEngineRequest(b *testing.B) {
	engine := actor.NewEngine()
	pid := engine.SpawnFunc(func(ctx *actor.Context) {
		switch ctx.Message().(type) {
		case []byte:
			ctx.Respond(ctx.Message())
		}
	}, "BenchmarkEngineRequest", actor.WithInboxSize(1024*8))
	payload := make([]byte, 128)
	wg := &sync.WaitGroup{}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = engine.Request(pid, payload, time.Second)
	}

	engine.Poison(pid, wg)
	wg.Wait()
}

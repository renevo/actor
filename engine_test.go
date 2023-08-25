package actor_test

import (
	"context"
	"log/slog"
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

	pid := engine.SpawnFunc(func(c *actor.Context) {
		t.Logf("PID %s Received %T", c.PID(), c.Message())

		switch msg := c.Message().(type) {
		case actor.Started:
		case actor.Stopped:
			t.Log("stopped!")
		case payload:
			if msg.data != 10 {
				panic("I failed to process this message")
			} else {
				t.Logf("finally processed all my messsages after borking %v.", msg.data)
			}
		}
	}, "foo", actor.WithRestartDelay(time.Millisecond*10), actor.WithTags("bar"))

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

	pid := engine.SpawnFunc(func(c *actor.Context) {
		switch c.Message().(type) {
		case actor.Initialized:
			slog.Info("init")
			wg.Add(1)
			init = true
		case actor.Started:
			slog.Info("start")
			is.True(init)
			started = true
		case int:
			slog.Info("msg")
			is.True(started)
			wg.Done()
		}
	}, "tst")

	engine.Send(context.Background(), pid, 1)
	wg.Wait()
	is.True(init)
	is.True(started)
}

func TestSendMsgRaceCon(t *testing.T) {
	engine := actor.NewEngine()
	wg := &sync.WaitGroup{}

	pid := engine.SpawnFunc(func(c *actor.Context) {
		msg := c.Message()
		if msg == nil {
			slog.Error("should never happen")
		}
	}, "test")

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
			}), "dummy", actor.WithTags(tag))
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
	pid := engine.SpawnFunc(func(c *actor.Context) {
		switch c.Message().(type) {
		case actor.Started:
			wg.Done()
		case actor.Stopped:
			atomic.AddInt32(&x, 1)
		}
	}, "foo")
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
	pid := engine.Spawn(newTickReceiver(wg), "test")
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
	}), "dummy")

	resp, err := engine.Request(pid, "foo", time.Millisecond)

	is.NoErr(err)
	is.Equal("bar", resp)
}

func BenchmarkEngineSend(b *testing.B) {
	engine := actor.NewEngine()
	pid := engine.SpawnFunc(func(*actor.Context) {}, "bench", actor.WithInboxSize(1024*8))
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
	}, "bench", actor.WithInboxSize(1024*8))
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

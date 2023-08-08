package actor_test

import (
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/renevo/actor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRestarts(t *testing.T) {
	engine := actor.NewEngine()
	wg := &sync.WaitGroup{}
	type payload struct {
		data int
	}

	wg.Add(1)
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
				wg.Done()
			}
		}
	}, "foo", actor.WithRestartDelay(time.Millisecond*10), actor.WithTags("bar"))

	engine.Send(pid, payload{1})
	engine.Send(pid, payload{2})
	engine.Send(pid, payload{10})
	engine.Poison(pid, wg)

	wg.Wait()
}

func TestProcessInitStartOrder(t *testing.T) {
	engine := actor.NewEngine()
	wg := &sync.WaitGroup{}
	var started, init bool

	pid := engine.SpawnFunc(func(c *actor.Context) {
		switch c.Message().(type) {
		case actor.Initialized:
			fmt.Println("init")
			wg.Add(1)
			init = true
		case actor.Started:
			fmt.Println("start")
			require.True(t, init)
			started = true
		case int:
			fmt.Println("msg")
			require.True(t, started)
			wg.Done()
		}
	}, "tst")

	engine.Send(pid, 1)
	wg.Wait()
}

func TestSendWithSender(t *testing.T) {
	engine := actor.NewEngine()
	sender := actor.NewPID("local", "sender")
	wg := &sync.WaitGroup{}

	wg.Add(1)
	pid := engine.SpawnFunc(func(c *actor.Context) {
		if _, ok := c.Message().(string); ok {
			assert.NotNil(t, c.Sender())
			assert.Equal(t, sender, c.Sender())
			wg.Done()
		}
	}, "test")

	engine.SendWithSender(pid, "data", sender)
	wg.Wait()
}

func TestSendMsgRaceCon(t *testing.T) {
	engine := actor.NewEngine()
	wg := &sync.WaitGroup{}

	pid := engine.SpawnFunc(func(c *actor.Context) {
		msg := c.Message()
		if msg == nil {
			fmt.Println("should never happen")
		}
	}, "test")

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			engine.Send(pid, []byte("f"))
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
			engine.Send(pid, 1)
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func TestPoisonWaitGroup(t *testing.T) {
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
	assert.Equal(t, int32(1), atomic.LoadInt32(&x))
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
	engine := actor.NewEngine()
	pid := engine.Spawn(NewTestReceiver(t, func(t *testing.T, ctx *actor.Context) {
		if msg, ok := ctx.Message().(string); ok {
			assert.Equal(t, "foo", msg)
			ctx.Respond("bar")
		}
	}), "dummy")

	resp, err := engine.Request(pid, "foo", time.Millisecond)

	assert.Nil(t, err)
	assert.Equal(t, "bar", resp)
}

func BenchmarkEngine(b *testing.B) {
	engine := actor.NewEngine()
	pid := engine.SpawnFunc(func(*actor.Context) {}, "bench", actor.WithInboxSize(1024*8))
	payload := make([]byte, 128)
	wg := &sync.WaitGroup{}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		engine.Send(pid, payload)
	}

	engine.Poison(pid, wg)
	wg.Wait()
}

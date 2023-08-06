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
	e := actor.NewEngine()
	wg := sync.WaitGroup{}
	type payload struct {
		data int
	}

	wg.Add(1)
	pid := e.SpawnFunc(func(c *actor.Context) {
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

	e.Send(pid, payload{1})
	e.Send(pid, payload{2})
	e.Send(pid, payload{10})
	e.Poison(pid, &wg)

	wg.Wait()
}

func TestProcessInitStartOrder(t *testing.T) {
	var (
		e             = actor.NewEngine()
		wg            = sync.WaitGroup{}
		started, init bool
	)
	pid := e.SpawnFunc(func(c *actor.Context) {
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
	e.Send(pid, 1)
	wg.Wait()
}

func TestSendWithSender(t *testing.T) {
	var (
		e      = actor.NewEngine()
		sender = actor.NewPID("local", "sender")
		wg     = sync.WaitGroup{}
	)
	wg.Add(1)

	pid := e.SpawnFunc(func(c *actor.Context) {
		if _, ok := c.Message().(string); ok {
			assert.NotNil(t, c.Sender())
			assert.Equal(t, sender, c.Sender())
			wg.Done()
		}
	}, "test")
	e.SendWithSender(pid, "data", sender)
	wg.Wait()
}

func TestSendMsgRaceCon(t *testing.T) {
	e := actor.NewEngine()
	wg := sync.WaitGroup{}

	pid := e.SpawnFunc(func(c *actor.Context) {
		msg := c.Message()
		if msg == nil {
			fmt.Println("should never happen")
		}
	}, "test")

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			e.Send(pid, []byte("f"))
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestSpawn(t *testing.T) {
	e := actor.NewEngine()
	wg := sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			tag := strconv.Itoa(i)
			pid := e.Spawn(NewTestReceiver(t, func(t *testing.T, ctx *actor.Context) {
			}), "dummy", actor.WithTags(tag))
			e.Send(pid, 1)
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func TestPoisonWaitGroup(t *testing.T) {
	var (
		e  = actor.NewEngine()
		wg = sync.WaitGroup{}
		x  = int32(0)
	)
	wg.Add(1)

	pid := e.SpawnFunc(func(c *actor.Context) {
		switch c.Message().(type) {
		case actor.Started:
			wg.Done()
		case actor.Stopped:
			atomic.AddInt32(&x, 1)
		}
	}, "foo")
	wg.Wait()

	pwg := &sync.WaitGroup{}
	e.Poison(pid, pwg)
	pwg.Wait()
	assert.Equal(t, int32(1), atomic.LoadInt32(&x))
}

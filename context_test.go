package actor_test

import (
	"sync"
	"testing"

	"github.com/matryer/is"
	"github.com/renevo/actor"
)

func TestSpawnChildPID(t *testing.T) {
	is := is.New(t)
	engine := actor.NewEngine()
	wg := &sync.WaitGroup{}
	childfn := func(c *actor.Context) {}
	expectedPID := actor.NewPID(actor.LocalAddress, "parent", "child")

	wg.Add(1)
	engine.SpawnFunc(func(c *actor.Context) {
		switch c.Message().(type) {
		case actor.Started:
			pid := c.SpawnFunc(childfn, "child")
			is.True(expectedPID.Equals(pid)) // child PID format mismatch
			wg.Done()
		case actor.Stopped:
		}
	}, "parent")

	wg.Wait()
}

func TestChild(t *testing.T) {
	is := is.New(t)

	engine := actor.NewEngine()
	wg := &sync.WaitGroup{}

	wg.Add(1)
	engine.SpawnFunc(func(c *actor.Context) {
		switch c.Message().(type) {
		case actor.Initialized:
			c.SpawnFunc(func(_ *actor.Context) {}, "child", actor.WithTags("1"))
			c.SpawnFunc(func(_ *actor.Context) {}, "child", actor.WithTags("2"))
			c.SpawnFunc(func(_ *actor.Context) {}, "child", actor.WithTags("3"))
		case actor.Started:
			is.Equal(3, len(c.Children())) // number of children match the amount spawned
			wg.Done()
		}
	}, "foo", actor.WithTags("bar", "baz"))
	wg.Wait()
}

func TestParent(t *testing.T) {
	is := is.New(t)

	engine := actor.NewEngine()
	wg := &sync.WaitGroup{}
	parent := actor.NewPID(actor.LocalAddress, "foo", "bar", "baz")

	wg.Add(1)

	childfn := func(c *actor.Context) {
		switch c.Message().(type) {
		case actor.Started:
			is.True(c.Parent().Equals(parent)) // parent should match
			is.Equal(len(c.Children()), 0)     // number of children should be zero
			wg.Done()
		}
	}

	engine.SpawnFunc(func(c *actor.Context) {
		switch c.Message().(type) {
		case actor.Started:
			c.SpawnFunc(childfn, "child")
		}
	}, "foo", actor.WithTags("bar", "baz"))

	wg.Wait()
}

func TestGetPID(t *testing.T) {
	is := is.New(t)

	engine := actor.NewEngine()
	wg := &sync.WaitGroup{}

	wg.Add(1)
	engine.SpawnFunc(func(c *actor.Context) {
		if _, ok := c.Message().(actor.Started); ok {
			pid := c.GetPID("foo", "bar", "baz")
			is.True(pid.Equals(c.PID()))
			wg.Done()
		}
	}, "foo", actor.WithTags("bar", "baz"))

	wg.Wait()
}

func TestSpawnChild(t *testing.T) {
	is := is.New(t)

	engine := actor.NewEngine()
	wg := &sync.WaitGroup{}
	deadletter := engine.GetPID("engine", "deadletter")

	wg.Add(1)
	childFunc := func(c *actor.Context) {
		switch c.Message().(type) {
		case actor.Stopped:
		}
	}

	pid := engine.SpawnFunc(func(ctx *actor.Context) {
		switch ctx.Message().(type) {
		case actor.Started:
			ctx.SpawnFunc(childFunc, "child", actor.WithMaxRestarts(0))
			wg.Done()
		}
	}, "parent", actor.WithMaxRestarts(0))

	wg.Wait()

	stopwg := &sync.WaitGroup{}
	engine.Poison(pid, stopwg)
	stopwg.Wait()

	is.Equal(deadletter, engine.GetPID("parent", "child"))
}

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

	childfn := func(ctx *actor.Context) {}
	expectedPID := actor.NewPID(actor.LocalAddress, "TestSpawnChildPID", "child")

	wg.Add(1)
	engine.SpawnFunc(func(ctx *actor.Context) {
		switch ctx.Message().(type) {
		case actor.Started:
			pid := ctx.SpawnFunc(childfn, "child")
			is.True(expectedPID.Equals(pid)) // child PID format mismatch
			wg.Done()
		case actor.Stopped:
		}
	}, "TestSpawnChildPID")

	wg.Wait()
}

func TestChild(t *testing.T) {
	is := is.New(t)

	engine := actor.NewEngine()
	wg := &sync.WaitGroup{}

	wg.Add(1)
	engine.SpawnFunc(func(ctx *actor.Context) {
		switch ctx.Message().(type) {
		case actor.Initialized:
			ctx.SpawnFunc(func(_ *actor.Context) {}, "child", actor.WithTags("1"))
			ctx.SpawnFunc(func(_ *actor.Context) {}, "child", actor.WithTags("2"))
			ctx.SpawnFunc(func(_ *actor.Context) {}, "child", actor.WithTags("3"))
		case actor.Started:
			is.Equal(3, len(ctx.Children())) // number of children match the amount spawned
			wg.Done()
		}
	}, "TestChild", actor.WithTags("bar", "baz"))
	wg.Wait()
}

func TestParent(t *testing.T) {
	is := is.New(t)

	engine := actor.NewEngine()
	wg := &sync.WaitGroup{}
	parent := actor.NewPID(actor.LocalAddress, "TestParent", "bar", "baz")

	wg.Add(1)
	childfn := func(ctx *actor.Context) {
		switch ctx.Message().(type) {
		case actor.Started:
			is.True(ctx.Parent().Equals(parent)) // parent should match
			is.Equal(len(ctx.Children()), 0)     // number of children should be zero
			wg.Done()
		}
	}

	engine.SpawnFunc(func(ctx *actor.Context) {
		switch ctx.Message().(type) {
		case actor.Started:
			ctx.SpawnFunc(childfn, "child")
		}
	}, "TestParent", actor.WithTags("bar", "baz"))

	wg.Wait()
}

func TestGetPID(t *testing.T) {
	is := is.New(t)

	engine := actor.NewEngine()
	wg := &sync.WaitGroup{}

	wg.Add(1)
	engine.SpawnFunc(func(ctx *actor.Context) {
		if _, ok := ctx.Message().(actor.Started); ok {
			pid := ctx.GetPID("TestGetPID", "bar", "baz")
			is.True(pid.Equals(ctx.PID()))
			wg.Done()
		}
	}, "TestGetPID", actor.WithTags("bar", "baz"))

	wg.Wait()
}

func TestSpawnChild(t *testing.T) {
	is := is.New(t)

	engine := actor.NewEngine()
	wg := &sync.WaitGroup{}
	deadletter := engine.GetPID("engine", "deadletter")

	wg.Add(1)
	childFunc := func(ctx *actor.Context) {
		switch ctx.Message().(type) {
		case actor.Stopped:
		}
	}

	pid := engine.SpawnFunc(func(ctx *actor.Context) {
		switch ctx.Message().(type) {
		case actor.Started:
			ctx.SpawnFunc(childFunc, "child", actor.WithMaxRestarts(0))
			wg.Done()
		}
	}, "TestSpawnChild", actor.WithMaxRestarts(0))

	wg.Wait()

	stopwg := &sync.WaitGroup{}
	engine.Poison(pid, stopwg)
	stopwg.Wait()

	is.Equal(deadletter, engine.GetPID("TestSpawnChild", "child"))
}

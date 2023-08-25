package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/renevo/actor"
)

const (
	dataDir = "./examples/data"
)

type PersistedActor struct {
	LastTime time.Time
}

func (p *PersistedActor) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case actor.Started:
		ctx.Log().Info("Startup", "last", p.LastTime)
	case actor.Stopped:
		ctx.Log().Info("Shutdown", "last", p.LastTime)
	case time.Time:
		p.LastTime = msg
	}
}

func main() {
	_ = os.MkdirAll(dataDir, os.ModePerm)

	engine := actor.NewEngine()
	persistenceMiddleware := actor.WithMiddleware(func(next actor.ReceiverFunc) actor.ReceiverFunc {
		return func(ctx *actor.Context) {
			switch ctx.Message().(type) {
			case actor.Initialized:
				// load on initialize
				if err := unmarshalActor(ctx.PID(), ctx.Receiver()); err != nil {
					ctx.Log().With("middleware", "persistence").Error("Failed to load actor", "err", err)
				}

				next(ctx)
			case actor.Started:
				// no need to save here
				next(ctx)

			default:
				next(ctx)

				// save after processing any other message
				if err := marshalActor(ctx.PID(), ctx.Receiver()); err != nil {
					ctx.Log().With("middleware", "persistence").Error("Failed to save actor", "err", err)
					return
				}
				ctx.Log().With("middleware", "persistence").Info("Saved State", "msg", reflect.TypeOf(ctx.Message()))
			}
		}
	})
	pid := engine.Spawn(&PersistedActor{}, "persisted", persistenceMiddleware)

	engine.Send(context.Background(), pid, time.Now())

	// shutdown
	engine.ShutdownAndWait()
}

func marshalActor(pid actor.PID, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("unable to marshal actor: %w", err)
	}

	if err := os.WriteFile(filepath.Join(dataDir, pid.ID+".json"), data, os.ModePerm); err != nil {
		return fmt.Errorf("unable to write actor data: %w", err)
	}
	return nil
}

func unmarshalActor(pid actor.PID, v any) error {
	data, err := os.ReadFile(filepath.Join(dataDir, pid.ID+".json"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("unable to read actor data: %w", err)
	}

	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("unable to unmarshal actor: %w", err)
	}

	return nil
}

package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/renevo/actor"
)

func main() {
	type contextKey string
	engineKey := contextKey("engine")
	middlewareKey := contextKey("middleware")

	// add a context value to all messages
	ctx := context.WithValue(context.Background(), engineKey, true)

	// create an engine and spawn a Receiver with middleware that injects a context.Value
	engine := actor.NewEngine(actor.WithContext(ctx))
	pid := engine.SpawnFunc(
		func(ctx *actor.Context) {
			switch ctx.Message().(type) {
			case string:
				hasEngineKey := ctx.Context().Value(engineKey)
				hasMiddlewareKey := ctx.Context().Value(middlewareKey)

				fmt.Printf("Engine: %t; Middleware: %t;\n", hasEngineKey, hasMiddlewareKey)
			}
		},
		"test",
		actor.WithMiddleware(func(next actor.ReceiverFunc) actor.ReceiverFunc {
			return func(ctx *actor.Context) {
				// inject a context value here
				next(ctx.WithContext(context.WithValue(ctx.Context(), middlewareKey, true)))
			}
		}))

	// this will overwrite the engine option context, so it is a good idea to use the one that was created originally
	engine.Send(ctx, pid, "hello")

	wg := &sync.WaitGroup{}
	engine.Poison(pid, wg)
	wg.Wait()
}

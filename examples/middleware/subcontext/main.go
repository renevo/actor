package main

import (
	"context"
	"fmt"

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
			hasEngineKey := ctx.Context().Value(engineKey)
			hasMiddlewareKey := ctx.Context().Value(middlewareKey)

			fmt.Printf("Engine: %t; Middleware: %t; Type: %T;\n", hasEngineKey, hasMiddlewareKey, ctx.Message())
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

	engine.ShutdownAndWait()
}

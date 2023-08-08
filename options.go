package actor

import (
	"context"
	"time"
)

const (
	defaultInboxSize   = 1024
	defaultMaxRestarts = 3
)

var (
	defaultRestartDelay = 500 * time.Millisecond
)

type Options struct {
	ID           []string
	Receiver     Receiver
	InboxSize    int
	BaseContext  context.Context
	MaxRestarts  int32
	RestartDelay time.Duration
	Middleware   []Middleware
}

type Option func(*Options)

func DefaultOptions(receiver Receiver) Options {
	return Options{
		Receiver:     receiver,
		InboxSize:    defaultInboxSize,
		MaxRestarts:  defaultMaxRestarts,
		RestartDelay: defaultRestartDelay,
		BaseContext:  context.Background(),
	}
}

func WithRestartDelay(d time.Duration) Option {
	return func(opts *Options) {
		opts.RestartDelay = d
	}
}

func WithInboxSize(size int) Option {
	return func(opts *Options) {
		opts.InboxSize = size
	}
}

func WithMaxRestarts(n int) Option {
	return func(opts *Options) {
		opts.MaxRestarts = int32(n)
	}
}

func WithContext(ctx context.Context) Option {
	return func(opts *Options) {
		if ctx == nil {
			ctx = context.Background()
		}
		opts.BaseContext = ctx
	}
}

func WithMiddleware(middleware ...Middleware) Option {
	return func(opts *Options) {
		opts.Middleware = append(opts.Middleware, middleware...)
	}
}

func WithTags(tags ...string) Option {
	return func(opts *Options) {
		opts.ID = append(opts.ID, tags...)
	}
}

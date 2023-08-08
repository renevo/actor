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
	Name         string
	Receiver     Receiver
	InboxSize    int
	MaxRestarts  int32
	RestartDelay time.Duration
	Middleware   []Middleware
	Tags         []string
	Context      context.Context
}

type Option func(*Options)

func DefaultOptions(receiver Receiver) Options {
	return Options{
		Receiver:     receiver,
		InboxSize:    defaultInboxSize,
		MaxRestarts:  defaultMaxRestarts,
		RestartDelay: defaultRestartDelay,
	}
}

func copyOptions(source *Options, receiver Receiver) *Options {
	return &Options{
		Receiver:     receiver,
		InboxSize:    source.InboxSize,
		MaxRestarts:  source.MaxRestarts,
		RestartDelay: source.RestartDelay,
		Context:      source.Context,
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

func WithMiddleware(middleware ...Middleware) Option {
	return func(opts *Options) {
		opts.Middleware = append(opts.Middleware, middleware...)
	}
}

func WithTags(tags ...string) Option {
	return func(opts *Options) {
		opts.Tags = append(opts.Tags, tags...)
	}
}

func WithContext(ctx context.Context) Option {
	return func(opt *Options) {
		opt.Context = ctx
	}
}

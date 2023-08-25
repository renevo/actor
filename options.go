package actor

import (
	"context"
	"log/slog"
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
	MaxRestarts  int
	RestartDelay time.Duration
	Middleware   []Middleware
	Tags         []string
	Context      context.Context
	Logger       *slog.Logger
}

type Option func(*Options)

func copyOptions(source *Options, receiver Receiver) *Options {
	return &Options{
		Receiver:     receiver,
		InboxSize:    source.InboxSize,
		MaxRestarts:  source.MaxRestarts,
		RestartDelay: source.RestartDelay,
		Context:      source.Context,
		Logger:       source.Logger,
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
		opts.MaxRestarts = n
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
		if ctx == nil {
			ctx = context.Background()
		}
		opt.Context = ctx
	}
}

func WithLogger(logger *slog.Logger) Option {
	return func(opt *Options) {
		if logger == nil {
			logger = slog.Default()
		}
		opt.Logger = logger
	}
}

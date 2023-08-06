package actor

import "sync"

type Receiver interface {
	Receive(ctx *Context)
}

type ReceiverFunc func(*Context)

func (f ReceiverFunc) Receive(ctx *Context) {
	f(ctx)
}

type poisonPill struct {
	wg *sync.WaitGroup
}

type Initialized struct{}
type Started struct{}
type Stopped struct{}

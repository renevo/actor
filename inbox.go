package actor

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrInboxClosed = errors.New("inbox closed")
)

type Envelope struct {
	To      PID
	From    PID
	Message any
	Context context.Context
}

type Inbox struct {
	box       chan *Envelope
	closeCh   chan struct{}
	closeOnce sync.Once
	startOnce sync.Once
	wg        sync.WaitGroup
}

func NewInbox(size int) *Inbox {
	in := &Inbox{}
	in.box = make(chan *Envelope, size)
	in.closeCh = make(chan struct{})
	return in
}

func (in *Inbox) Process(proc Processor) {
	in.startOnce.Do(func() {
		in.wg.Add(1)

		go func() {
			for env := range in.box {
				proc.Process(env)
			}
			in.wg.Done()
		}()
	})
}

func (in *Inbox) Deliver(env *Envelope) error {
	select {
	case <-in.closeCh:
		return ErrInboxClosed

	default:
		in.box <- env
	}

	return nil
}

func (in *Inbox) Close() {
	in.closeOnce.Do(func() {
		close(in.closeCh)
		close(in.box)
	})
}

func (in *Inbox) Drain() {
	in.Close()
	in.wg.Wait()
}

package actor

import "time"

type Repeater struct {
	engine   *Engine
	self     PID
	target   PID
	msg      any
	interval time.Duration
	stopCh   chan struct{}
}

func (r Repeater) start() {
	go func() {
		t := time.NewTicker(r.interval)
		defer t.Stop()

		for {
			select {
			case <-t.C:
				r.engine.SendWithSender(r.target, r.msg, r.self)

			case <-r.stopCh:
				return
			}
		}
	}()
}

// Stop the Repeater. This will panic if called more than once
func (r Repeater) Stop() {
	close(r.stopCh)
}

// SendRepeat will send the given message to the given PID each given interval.
// It will return a Repeater struct that can stop the repeating message by calling Stop().
func (e *Engine) SendRepeat(to PID, msg any, interval time.Duration) Repeater {
	repeater := Repeater{
		engine:   e,
		self:     e.pid,
		target:   to,
		interval: interval,
		msg:      msg,
		stopCh:   make(chan struct{}, 1),
	}
	repeater.start()

	return repeater
}

// SendRepeat will send the given message to the given PID each given interval.
// It will return a Repeater struct that can stop the repeating message by calling Stop().
func (c *Context) SendRepeat(to PID, msg any, interval time.Duration) Repeater {
	repeater := Repeater{
		engine:   c.engine,
		self:     c.pid,
		target:   to,
		interval: interval,
		msg:      msg,
		stopCh:   make(chan struct{}, 1),
	}
	repeater.start()
	return repeater
}

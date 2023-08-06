package actor

type Engine struct {
}

func (e *Engine) Deliver(env Envelope) {
	// find the inbox and deliver it, otherwise, deliver it to the deadletter
}

package actor

type Context struct {
	pid           PID
	sender        PID
	engine        *Engine
	receiver      Receiver
	message       any
	parentContext *Context
	children      *safemap[string, PID]
}

func newContext(e *Engine, pid PID) *Context {
	return &Context{
		engine:   e,
		pid:      pid,
		children: newMap[string, PID](),
	}
}

func (c *Context) Reciever() Receiver {
	return c.receiver
}

// Send message to another actor (core functionality)
func (c *Context) Send(to PID, msg any) {
	c.engine.SendWithSender(to, msg, c.pid)
}

// Forward the current message to another PID
func (c *Context) Forward(to PID) {
	c.engine.SendWithSender(to, c.message, c.pid)
}

func (c *Context) GetPID(id ...string) PID {
	return c.engine.GetPID(id...)
}

func (c *Context) PID() PID {
	return c.pid
}

func (c *Context) Sender() PID {
	return c.sender
}

func (c *Context) Engine() *Engine {
	return c.engine
}

func (c *Context) Message() any {
	return c.message
}

func (c *Context) Parent() PID {
	if c.parentContext != nil {
		return c.parentContext.pid
	}

	return c.engine.pid
}

func (c *Context) Child(id string) (PID, bool) {
	return c.children.Get(id)
}

func (c *Context) Children() []PID {
	pids := make([]PID, c.children.Len())
	i := 0
	c.children.ForEach(func(_ string, child PID) {
		pids[i] = child
		i++
	})
	return pids
}

// Receiver returns the underlying receiver of this Context.
func (c *Context) Receiver() Receiver {
	return c.receiver
}

// Spawn a new actor (core functionality)
func (c *Context) Spawn(receiver Receiver, name string, opts ...Option) PID {
	options := DefaultOptions(receiver)
	options.ID = []string{c.PID().ID, name}
	for _, opt := range opts {
		opt(&options)
	}
	proc := newProcessor(c.engine, options)
	proc.context.parentContext = c
	pid := c.engine.SpawnProcessor(proc)
	c.children.Set(pid.ID, pid)

	return proc.PID()
}

func (c *Context) SpawnFunc(fn ReceiverFunc, name string, opts ...Option) PID {
	return c.Spawn(fn, name, opts...)
}

// TODO: Add request/response

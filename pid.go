package actor

import "strings"

type PID struct {
	Address string
	ID      string
}

func NewPID(address string, name string, tags ...string) PID {
	if address == "" {
		address = LocalAddress
	}

	return PID{
		Address: address,
		ID:      strings.Join(append([]string{name}, tags...), pidSeparator),
	}
}

func (p PID) Equals(pid PID) bool {
	return pid.Address == p.Address && pid.ID == p.ID
}

func (p PID) String() string {
	return p.Address + pidSeparator + p.ID
}

func (p PID) Child(name string, tags ...string) PID {
	return NewPID(p.Address, p.ID, append([]string{name}, tags...)...)
}

func (p PID) IsZero() bool {
	return p.Address == "" && p.ID == ""
}

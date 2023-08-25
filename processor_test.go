package actor_test

import (
	"testing"

	"github.com/matryer/is"
	"github.com/renevo/actor"
)

func TestPID(t *testing.T) {
	is := is.New(t)

	pid := actor.NewPID(actor.LocalAddress, "test", "tag1", "tag2")
	is.Equal(pid.String(), "local.test.tag1.tag2")
	is.Equal(pid.ID, "test.tag1.tag2")
	is.Equal(pid.Address, "local")

	cpid := pid.Child("child")
	is.Equal(cpid.String(), "local.test.tag1.tag2.child")
	is.Equal(cpid.ID, "test.tag1.tag2.child")
	is.Equal(cpid.Address, "local")
}

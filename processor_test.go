package actor_test

import (
	"testing"

	"github.com/renevo/actor"
	"github.com/stretchr/testify/assert"
)

func TestPID(t *testing.T) {
	pid := actor.NewPID(actor.LocalAddress, "test", "tag1", "tag2")
	assert.Equal(t, pid.String(), "local.test.tag1.tag2")
	assert.Equal(t, pid.ID, "test.tag1.tag2")
	assert.Equal(t, pid.Address, "local")

	cpid := pid.Child("child")
	assert.Equal(t, cpid.String(), "local.test.tag1.tag2.child")
	assert.Equal(t, cpid.ID, "test.tag1.tag2.child")
	assert.Equal(t, cpid.Address, "local")
}

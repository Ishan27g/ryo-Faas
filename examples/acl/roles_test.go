package acl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRole(t *testing.T) {

	admin := NewRole("admin", mostAccess)
	user1 := NewRole("user1", midAccess1, midAccess3)
	user2 := NewRole("user2", midAccess2, midAccess3)
	guest := NewRole("guest", leastAccess)

	admin.AddChild(user1, user2, guest)
	user1.AddChild(guest)
	user2.AddChild(guest)

	assert.True(t, admin.CanAccess(leastAccess))
	assert.True(t, admin.CanAccess(midAccess1))
	assert.True(t, admin.CanAccess(midAccess2))
	assert.True(t, admin.CanAccess(midAccess3))
	assert.True(t, admin.CanAccess(mostAccess))

	assert.True(t, user1.CanAccess(midAccess1))
	assert.True(t, user1.CanAccess(midAccess3))
	assert.True(t, user1.CanAccess(leastAccess))
	assert.False(t, user1.CanAccess(midAccess2))
	assert.False(t, user1.CanAccess(mostAccess))

	assert.True(t, user2.CanAccess(midAccess2))
	assert.True(t, user2.CanAccess(midAccess3))
	assert.True(t, user2.CanAccess(leastAccess))
	assert.False(t, user2.CanAccess(midAccess1))
	assert.False(t, user2.CanAccess(mostAccess))

	assert.True(t, guest.CanAccess(leastAccess))
	assert.False(t, guest.CanAccess(midAccess3))
	assert.False(t, guest.CanAccess(midAccess2))
	assert.False(t, guest.CanAccess(midAccess1))
	assert.False(t, guest.CanAccess(mostAccess))

}

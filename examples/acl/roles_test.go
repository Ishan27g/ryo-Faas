package acl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var userAccess1 = Permission{"write-resource-1"}
var userAccess2 = Permission{"write-resource-2"}
var userAccess3 = Permission{"write-resource-3"}
var adminAccess = Permission{"delete-resources"}

func TestNewRole(t *testing.T) {

	admin := NewRole("admin", adminAccess)
	user1 := NewRole("user1", userAccess1, userAccess3)
	user2 := NewRole("user2", userAccess2, userAccess3)
	guest := NewRole("guest", guestAccess)

	admin.AddChild(user1, user2, guest)
	user1.AddChild(guest)
	user2.AddChild(guest)

	assert.True(t, admin.CanAccess(guestAccess))
	assert.True(t, admin.CanAccess(userAccess1))
	assert.True(t, admin.CanAccess(userAccess2))
	assert.True(t, admin.CanAccess(userAccess3))
	assert.True(t, admin.CanAccess(adminAccess))

	assert.True(t, user1.CanAccess(userAccess1))
	assert.True(t, user1.CanAccess(userAccess3))
	assert.True(t, user1.CanAccess(guestAccess))
	assert.False(t, user1.CanAccess(userAccess2))
	assert.False(t, user1.CanAccess(adminAccess))

	assert.True(t, user2.CanAccess(userAccess2))
	assert.True(t, user2.CanAccess(userAccess3))
	assert.True(t, user2.CanAccess(guestAccess))
	assert.False(t, user2.CanAccess(userAccess1))
	assert.False(t, user2.CanAccess(adminAccess))

	assert.True(t, guest.CanAccess(guestAccess))
	assert.False(t, guest.CanAccess(userAccess3))
	assert.False(t, guest.CanAccess(userAccess2))
	assert.False(t, guest.CanAccess(userAccess1))
	assert.False(t, guest.CanAccess(adminAccess))

}

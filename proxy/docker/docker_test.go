package docker

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDocker_Start(t *testing.T) {

	err := New().StartNats()
	assert.NoError(t, err)

	err = New().StartProxy()
	assert.NoError(t, err)

	<-time.After(1 * time.Second)
	err = New().StartDatabase()
	assert.NoError(t, err)
	//
	//<-time.After(1 * time.Second)
	err = New().StartAgent("1")
	assert.NoError(t, err)

}

func TestDocker_Stop(t *testing.T) {
	d := New()
	assert.NotNil(t, d)
	err := d.StopNats()
	assert.NoError(t, err)

	err = d.StopProxy()
	assert.NoError(t, err)

	<-time.After(1 * time.Second)
	err = d.StopDatabase()
	assert.NoError(t, err)
	//
	//<-time.After(1 * time.Second)
	err = d.StopAgent("1")
	assert.NoError(t, err)

}

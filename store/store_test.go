package store

import (
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/DavidGamba/dgtools/run"
	"github.com/Ishan27g/ryo-Faas/pkg/shell"
	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	var dbShell shell.Shell

	os.Setenv("NATS", "nats://localhost:4222")

	err := os.Chdir("../database")
	assert.NoError(t, err)

	_, pw := io.Pipe()
	defer pw.Close()

	shOpts := []shell.Option{shell.WithOutput(pw), shell.WithCmd(run.CMD("go", "run", "main.go"))}
	dbShell = shell.New(shOpts...)
	defer func() {
		dbShell.Kill()
	}()
	if !dbShell.Run() {
		t.Error("cannot start database")
	}
	var gets, updates, creates, deletes = 0, 0, 0, 0
	docStore := Get("payments")
	docStore.On(DocumentGET, func(document Doc) {
		fmt.Println("sub-get:", document.Data)
		gets++
	})
	docStore.On(DocumentUPDATE, func(document Doc) {
		fmt.Println("sub-update:", document.Data)
		updates++
	})
	docStore.On(DocumentCREATE, func(document Doc) {
		fmt.Println("sub-create:", document.Data)
		creates++
	})
	docStore.On(DocumentDELETE, func(document Doc) {
		fmt.Println("sub-delete:", document.Data)
		deletes++
	})
	<-time.After(1 * time.Second)
	// data to add
	data := map[string]interface{}{
		"from":   "bob",
		"to":     "alice",
		"amount": 42,
	}

	// add a new `payment` to the db
	id := docStore.Create("", data)
	assert.NotEqual(t, "", id)
	// get it from the db
	dataReturned := docStore.Get(id)
	assert.NotEmpty(t, dataReturned)
	assert.True(t, docStore.Update(id, data))
	assert.True(t, docStore.Update(id, data))
	assert.True(t, docStore.Delete(id))

	<-time.After(1 * time.Second)

	assert.Equal(t, 1, creates)
	assert.Equal(t, 4, gets) // 1 for each get,update,delete
	assert.Equal(t, 2, updates)
	assert.Equal(t, 1, deletes)

}

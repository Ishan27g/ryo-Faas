package store

import (
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/DavidGamba/dgtools/run"
	"github.com/Ishan27g/ryo-Faas/shell"
	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	var dbShell shell.Shell

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

	docStore := Get("payments")
	docStore.On("get", func(document Doc) {
		fmt.Println("SUB:", document.Data.Value)
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

	fmt.Println(dataReturned)

	//// dataReturned == data
	//fmt.Println(dataReturned)
	//
	//// update some field
	//data["amount"] = 43
	//docStore.Update(id, data)
	//
	//// delete it
	//docStore.Delete(id)
}

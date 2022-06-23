package cmd

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var exampleHttp = `
package main

import (
	"net/http"

	FuncFw "github.com/Ishan27g/ryo-Faas/funcFw"
)

func Example(w http.ResponseWriter, r *http.Request) {}

`
var exampleGin = `
package main

import (
	"github.com/gin-gonic/gin"
	FuncFw "github.com/Ishan27g/ryo-Faas/funcFw"
)
func Example(c *gin.Context) {
	c.JSON(http.StatusOK, nil)
}

`
var test = func(t *testing.T, fname []byte) (bool, bool) {

	file, err := ioutil.TempFile("./", "")
	defer file.Close()
	defer os.Remove(file.Name())
	assert.NoError(t, err)
	_, err = file.Write(fname)
	assert.NoError(t, err)
	return validate(file.Name(), "")
}

func Test_validate(t *testing.T) {

	valid, isStdHttp := test(t, []byte(exampleGin))
	assert.True(t, valid)
	assert.False(t, isStdHttp)
	valid, isStdHttp = test(t, []byte(exampleHttp))
	assert.True(t, valid)
	assert.True(t, isStdHttp)
}

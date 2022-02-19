package registry

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAstGen(t *testing.T) {

	pathToFnFw = path() + FnFw

	packageName := "method1"
	entrypoint := "Method1"

	dotGo, err := rewriteDeployDotGo(packageName, entrypoint)
	if err != nil {
		return
	}
	assert.NoError(t, err)
	fmt.Println(dotGo)
}

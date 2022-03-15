package registry

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAstGen(t *testing.T) {

	pathToDeployment = path() + FnFw

	var exports = []function{
		{
			pkgName:    "method1",
			entrypoint: "Method1",
		},
		{
			pkgName:    "methodOtel",
			entrypoint: "MethodWithOtel",
		},
	}

	// change relative path
	pathToDeployment = path() + FnFw
	PathToFns = pathToDeployment + "functions/"
	modFile = func() string {
		return path() + "/tmplt/template.go"
	}

	dotGo, err := rewriteDeployDotGo(exports...)
	assert.NoError(t, err)
	fmt.Println(dotGo)
}

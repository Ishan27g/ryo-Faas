package helloWorld

import (
	"fmt"
	"net/http"

	FuncFw "github.com/Ishan27g/ryo-Faas/funcFw"
)

type srv struct {
	val string
}

func init() {
	// inject ctx in funcfw
	s := &srv{"some data"}
	FuncFw.InjectCtx(s)
}

func Hello(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Hello hit....")

	// get ctx from funcfw
	s := FuncFw.ExtractCtx[*srv]()
	if s != nil {
		w.Write([]byte(s.val))
	}
	w.WriteHeader(http.StatusAccepted)
}

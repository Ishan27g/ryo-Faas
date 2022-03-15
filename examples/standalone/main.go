package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Ishan27g/ryo-Faas/agent/registry/deploy"
)

func Method2(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprint(w, "Accepted at method 2 ..."+"\n")
}

func Method1(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprint(w, "Accepted at method 1 ..."+"\n")
}
func main() {
	FuncFw.Export("Method2", "/method2", Method2)
	FuncFw.Export("Method1", "/method1", Method1)

	FuncFw.Start("9999")

	<-time.After(10 * time.Second)
	FuncFw.Stop()
}

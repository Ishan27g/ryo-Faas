package main

import (
	"fmt"
	"net/http"
	"time"

	FuncFw "github.com/Ishan27g/ryo-Faas/funcFw"
	"github.com/Ishan27g/ryo-Faas/store"
)

func Method2(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprint(w, "Accepted at method 2 ..."+"\n")
}

func Method1(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprint(w, "Accepted at method 1 ..."+"\n")
}
func Events() store.StoreEvents {
	return store.StoreEvents{
		OnCreate: []store.EventCb{},
		OnGet:    []store.EventCb{},
		OnUpdate: []store.EventCb{},
		OnDelete: []store.EventCb{},
	}
}
func main() {

	Events().Apply()

	FuncFw.Export.Http("Method2", "/method2", Method2)
	FuncFw.Export.Http("Method1", "/method1", Method1)

	FuncFw.Start("9999")

	<-time.After(10 * time.Second)
	FuncFw.Stop()
}

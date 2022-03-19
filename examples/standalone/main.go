package main

import (
	"fmt"
	"net/http"
	"time"

	FuncFw "github.com/Ishan27g/ryo-Faas/funcFw"
	"github.com/Ishan27g/ryo-Faas/store"
)

func Method3(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Processing method 3 ...")
	<-time.After(2 * time.Second)
	fmt.Println("Processing method 3 done")
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprint(w, "Accepted at method 3 ..."+"\n")
}
func Method2(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprint(w, "Accepted at method 2 ..."+"\n")
}

func Method1(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprint(w, "Accepted at method 1 ..."+"\n")
}

func GenericCb() func(document store.NatsDoc) {
	return func(document store.NatsDoc) {
		fmt.Println(document.Table() + " " + document.Id() + " ---- at GenericCb()")
		// ...
	}
}
func main() {

	FuncFw.Export.Http("Method2", "/method2", Method2)
	FuncFw.Export.Http("Method1", "/method1", Method1)

	FuncFw.Export.HttpAsync("Async", "/async", Method3)

	FuncFw.Export.NatsAsync("AsyncNats", "/asyncNats", Method3)

	FuncFw.Export.EventsFor("payments").On(store.DocumentCREATE, GenericCb())
	FuncFw.Export.EventsFor("bills").On(store.DocumentCREATE, GenericCb())
	FuncFw.Export.EventsFor("payments").OnIds(store.DocumentGET, GenericCb(), "some-known-id")
	FuncFw.Export.EventsFor("bills").OnIds(store.DocumentGET, GenericCb(), "some-known-id")
	FuncFw.Export.EventsFor("payments").OnIds(store.DocumentUPDATE, GenericCb(), "some-known-id")

	FuncFw.Start("9999")

	<-time.After(10000 * time.Second)
	FuncFw.Stop()
}

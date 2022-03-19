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

	FuncFw.EventsForTable("payments").OnCreate(GenericCb())
	FuncFw.EventsForTable("bills").OnUpdate(GenericCb())
	FuncFw.EventsForTable("payments").OnUpdateIds(GenericCb(), "some-known-id")
	FuncFw.EventsForTable("bills").OnDeleteIds(GenericCb(), "some-known-id")
	FuncFw.EventsForTable("payments").OnGetIds(GenericCb(), "some-known-id")

	FuncFw.Export.Http("Method2", "/method2", Method2)
	FuncFw.Export.Http("Method1", "/method1", Method1)

	FuncFw.Export.HttpAsync("TodoAsync", "/todoAsync", Method3)

	FuncFw.Start("9999")

	<-time.After(10000 * time.Second)
	FuncFw.Stop()
}

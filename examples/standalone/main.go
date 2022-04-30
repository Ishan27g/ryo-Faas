package main

import (
	"fmt"
	"net/http"
	"time"

	FuncFw "github.com/Ishan27g/ryo-Faas/funcFw"
	"github.com/Ishan27g/ryo-Faas/store"
	"github.com/gin-gonic/gin"
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
func cb(document store.Doc) {
	fmt.Println(document.CreatedAt + " " + document.Id + " ---- at GenericCb()")
}
func Test(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.RequestURI)
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprint(w, "Accepted at Test ..."+"\n")
}
func testGin(c *gin.Context) {
	c.JSON(http.StatusOK, nil)
}
func main() {

	FuncFw.Export.Http("testMethod", "/test", Test)
	FuncFw.Export.HttpGin("testMethod", "/testgin", testGin)

	//FuncFw.Export.Http("Method2", "/method2", Method2)
	//FuncFw.Export.Http("Method1", "/method1", Method1)
	//
	//FuncFw.Export.HttpAsync("Async", "/async", Method3)
	//
	//FuncFw.Export.NatsAsync("AsyncNats", "/asyncNats", Method3)
	//
	//FuncFw.Export.EventsFor("payments").On(store.DocumentCREATE, cb)
	//FuncFw.Export.EventsFor("bills").On(store.DocumentCREATE, cb)
	//FuncFw.Export.EventsFor("payments").OnIds(store.DocumentGET, cb, "some-known-id")
	//FuncFw.Export.EventsFor("bills").OnIds(store.DocumentGET, cb, "some-known-id")
	//FuncFw.Export.EventsFor("payments").OnIds(store.DocumentUPDATE, cb, "some-known-id")

	FuncFw.Start("9999")

	<-time.After(10000 * time.Second)
	FuncFw.Stop()
}

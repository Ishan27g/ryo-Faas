package main

import (
	"fmt"
	"net/http"
	"time"

	FuncFw "github.com/Ishan27g/ryo-Faas/funcFw"
)

func Method2(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprint(w, "Accepted at method 2 ..."+"\n")
}

func Method1(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprint(w, "Accepted at method 1 ..."+"\n")
}
func Events() FuncFw.StoreEvents {
	return FuncFw.StoreEvents{}
}
func main() {

	FuncFw.Export.Events(FuncFw.StoreEvents{
		OnCreate: FuncFw.Events{},
		OnGet:    FuncFw.Events{},
		OnUpdate: FuncFw.Events{},
		OnDelete: FuncFw.Events{},
	})

	FuncFw.Export.Http("Method2", "/method2", Method2)
	FuncFw.Export.Http("Method1", "/method1", Method1)

	FuncFw.Start("9999")

	<-time.After(10 * time.Second)
	FuncFw.Stop()
}

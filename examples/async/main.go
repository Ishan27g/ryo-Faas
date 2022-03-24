package async

import (
	"fmt"
	"net/http"
	"time"
)

// MethodAsync is a long-running job
func MethodAsync(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Processing ...")
	<-time.After(5 * time.Second)

	fmt.Println("Still processing ...")
	<-time.After(5 * time.Second)

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Async processing done ..."+"\n")
}

//func main() {
//
//	FuncFw.Export.HttpAsync("Async-process", "/process", MethodAsync)
//
//	FuncFw.Start("9999")
//
//	<-time.After(10000 * time.Second)
//	FuncFw.Stop()
//}

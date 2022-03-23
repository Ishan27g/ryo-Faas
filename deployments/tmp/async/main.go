package async

import (
	"fmt"
	"net/http"
	"time"
)

// MethodAsync is a long-running job
func MethodAsync(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Processing method 3 ...")

	<-time.After(15 * time.Second)
	fmt.Println("Processing method 3 done")

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Async processing done at method 3 ..."+"\n")
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

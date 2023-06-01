package async

import (
	"fmt"
	"net/http"
	"time"

	FuncFw "github.com/Ishan27g/ryo-Faas/funcFw"
)

type invocation struct {
	count int
}

func init() {
	count := 0
	FuncFw.InjectCtx(&invocation{count})
}

// MethodAsync is a long-running job
func MethodAsync(w http.ResponseWriter, r *http.Request) {
	i := FuncFw.ExtractCtx[*invocation]()
	i.count = (i.count) + 1

	fmt.Println(fmt.Sprintf("Async process %d started ...", i.count))

	fmt.Println("Processing ...")
	<-time.After(2 * time.Second)

	fmt.Println("Still processing ...")
	<-time.After(2 * time.Second)

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, fmt.Sprintf("Async process %d done ..."+"\n", i.count))
}

func main() {
	//
	//FuncFw.Export.HttpAsync("Async-process", "/functions/methodasync", MethodAsync)
	//
	//FuncFw.Start("9999", "")
	//
	//<-time.After(10000 * time.Second)
	//FuncFw.Stop()
}

package example

import (
	"fmt"
	"net/http"
)

// MethodOk
func Method2(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
	fmt.Println("method2 hit....")
	fmt.Fprint(w, "Accepted at method 2 ..."+"\n")
}

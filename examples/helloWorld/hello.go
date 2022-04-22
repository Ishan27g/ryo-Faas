package helloWorld

import (
	"fmt"
	"net/http"
)

func Hello(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
	fmt.Println("Hello hit....")
	fmt.Fprint(w, "Hello ..."+"\n")
}

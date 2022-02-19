package method1

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type request struct {
	Data string `json:"data"`
}

// MethodOk
func Method1(w http.ResponseWriter, r *http.Request) {
	var p request
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	status, rsp := sendHttpTo(p.Data)
	if status == -1 {
		http.Error(w, "Error reaching "+p.Data, http.StatusBadRequest)
		return
	}
	w.WriteHeader(status)
	fmt.Fprint(w, rsp+"\nAccepted at Method 1 ..."+"\n")
}

func sendHttpTo(url string) (int, string) {
	fmt.Println("getting from url", url)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err.Error())
		return -1, ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		return -1, ""
	}
	fmt.Println("RESP - ", string(body))
	return resp.StatusCode, string(body)
}

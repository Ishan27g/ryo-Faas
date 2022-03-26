package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Ishan27g/ryo-Faas/examples/acl"
)

func main2() {
	var payload acl.RoleJson
	payload.Id = "admin"
	payload.Permissions = []string{"ok", "okay"}

	marshal, err := json.Marshal(payload)
	if err != nil {
		return
	}

	send("http://localhost:9999/add", marshal)
	<-time.After(1 * time.Second)
	payload.Permissions = []string{"ok"}
	marshal, err = json.Marshal(payload)
	if err != nil {
		return
	}
	send("http://localhost:9999/gwt", marshal)

	payload.Permissions = []string{"ok1"}
	marshal, err = json.Marshal(payload)
	if err != nil {
		return
	}
	send("http://localhost:9999/gwt", marshal)
}

func send(url string, body []byte) {

	request, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return
	}

	client := http.Client{Timeout: 3 * time.Second}

	res, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	body, err = ioutil.ReadAll(res.Body)
	_ = res.Body.Close()
	fmt.Println("response is ", string(body))
}

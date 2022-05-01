package transport

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func SendHttp(method string, url string, body []byte) ([]byte, int) {
	request, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, -1
	}

	client := http.Client{Timeout: 3 * time.Second}

	res, err := client.Do(request)
	if err != nil {
		fmt.Println(err.Error())
		return nil, -1
	}
	body, err = ioutil.ReadAll(res.Body)
	_ = res.Body.Close()
	fmt.Println("response is ", string(body))
	return body, res.StatusCode
}

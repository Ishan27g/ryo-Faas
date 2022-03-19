package FuncFw

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/Ishan27g/ryo-Faas/store"
	"github.com/Ishan27g/ryo-Faas/transport"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
)

func TestNatsSimple(t *testing.T) {
	var ok = false
	transport.NatsSubscribe(store.DocumentCREATE, func(msg *nats.Msg) {
		fmt.Println("[SUB]", string(msg.Data))
		ok = true
	})
	transport.NatsPublish(store.DocumentCREATE, "nice", nil)
	<-time.After(2 * time.Second)
	if !ok {
		t.Error("no sub")
	}
}

func TestNatsJson(t *testing.T) {
	var ok = false

	var sendStruct = NewAsyncNats("any", "")

	go transport.NatsSubscribeJson(sendStruct.getSubj(), func(msg *transport.AsyncNats) {
		req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(msg.Req)))
		assert.NoError(t, err)
		assert.Equal(t, req.Method, "POST")
		assert.Equal(t, req.RequestURI, "/anything")
		assert.Equal(t, req.Header.Get("X-Custom-Header"), "myvalue")
		assert.Equal(t, req.Header.Get("Content-Type"), "application/json")
		body, err := ioutil.ReadAll(req.Body)
		assert.NoError(t, err)
		assert.Equal(t, `{"title":"Buy cheese and bread for breakfast."}`, string(body))
		ok = true
	})
	<-time.After(2 * time.Second)
	var jsonStr = []byte(`{"title":"Buy cheese and bread for breakfast."}`)
	sendStruct.req, _ = http.NewRequest("POST", "/anything", bytes.NewBuffer(jsonStr))
	sendStruct.req.Header.Set("X-Custom-Header", "myvalue")
	sendStruct.req.Header.Set("Content-Type", "application/json")
	var b = &bytes.Buffer{}
	err := sendStruct.req.Write(b)
	assert.NoError(t, err)
	if !transport.NatsPublishJson(sendStruct.getSubj(), transport.AsyncNats{
		Callback:   sendStruct.callback,
		Entrypoint: sendStruct.entrypoint,
		Req:        b.Bytes(),
	}, nil) {
		t.Error("could not publish")
		return
	}
	<-time.After(3 * time.Second)
	if !ok {
		t.Error("no sub")
	}
}

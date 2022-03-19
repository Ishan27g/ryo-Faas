package FuncFw

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Ishan27g/ryo-Faas/store"
	"github.com/Ishan27g/ryo-Faas/transport"
	"github.com/gin-gonic/gin"
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

var (
	mockCbServerPort = ":9090"
	mockCbServerUrl  = "/any"
	callBackAddress  = "http://localhost" + mockCbServerPort + mockCbServerUrl

	serverCallbackHit = false
)

func mockCallBackServer(ctx context.Context) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Any(mockCbServerUrl, func(c *gin.Context) {
		serverCallbackHit = true
		fmt.Println("serverCallbackHit", c.Request.URL.RequestURI(), " from ", c.Request.RemoteAddr)
		c.JSON(http.StatusOK, nil)
	})
	httpSrv := &http.Server{Addr: mockCbServerPort, Handler: r}
	go func() {
		fmt.Println("mockCallBackServer started on " + callBackAddress)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("HTTP", err.Error())
		}
	}()
	<-ctx.Done()
	cx, can := context.WithTimeout(context.Background(), 2*time.Second)
	defer can()
	if err := httpSrv.Shutdown(cx); err != nil {
		fmt.Println("Http-Shutdown " + err.Error())
	}
}
func mockIncomingRequest(t *testing.T, cb func(w http.ResponseWriter, r *http.Request)) {

	ww := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "", nil)
	req.Header.Set("X-Callback-Url", callBackAddress)
	cb(ww, req)
	assert.Equal(t, http.StatusAccepted, ww.Result().StatusCode)
}
func TestNatsHttpFunction(t *testing.T) {
	var method2 = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		fmt.Println("method2 hit....")
		fmt.Fprint(w, "Accepted at method 2 ..."+"\n")
	}

	Export.NatsAsync("testAsync", "/testAsync", method2)

	for _, na := range Export.GetHttpAsyncNats() {
		an := NewAsyncNats(na.Entrypoint, "")
		an.SubscribeAsync(na.HttpFn)
	}
	<-time.After(3 * time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go mockCallBackServer(ctx)
	for _, na := range Export.GetHttpAsyncNats() {
		an := NewAsyncNats(na.Entrypoint, "")
		mockIncomingRequest(t, an.HandleAsyncNats)
	}

	<-time.After(3 * time.Second)
	assert.Equal(t, true, serverCallbackHit)

}

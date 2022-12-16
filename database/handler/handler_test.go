package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	client "github.com/Ishan27g/ryo-Faas/database/client"
	database "github.com/Ishan27g/ryo-Faas/database/db"
	deploy "github.com/Ishan27g/ryo-Faas/pkg/proto"
	"github.com/Ishan27g/ryo-Faas/store"

	"github.com/Ishan27g/ryo-Faas/pkg/transport"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var id string
var entity database.Entity
var data = map[string]interface{}{
	"Num":  1,
	"From": 300,
	"To":   302,
}
var table1 = map[string]interface{}{
	"Table": "ok",
	"Data":  data,
}
var table2 = map[string]interface{}{
	"Table": "ok2",
	"Data":  data,
}

func Test_Grpc(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var handler = GetHandler()

	config := []transport.Config{transport.WithRpcPort(":5001"), transport.WithDatabaseServer(&handler.Rpc)}
	transport.Init(ctx, config...).Start()

	<-time.After(5 * time.Second)

	c := client.Connect("localhost:5001")

	d, err := json.Marshal(table1["Data"])
	assert.NoError(t, err)

	id, err := c.New(ctx, &deploy.Documents{Document: []*deploy.Document{{Table: table1["Table"].(string), Id: "", Data: d}}})
	docId := id.GetId()[0]
	assert.NoError(t, err)

	doc, err := c.Get(ctx, &deploy.Ids{Id: id.Id})

	docs := store.ToDocs(doc, "Table")

	for _, d := range docs {
		assert.Equal(t, docId, d.Id)
		assert.NotNil(t, d.Data["Num"])
		assert.NotNil(t, d.Data["From"])
		assert.NotNil(t, d.Data["To"])
	}

	deletedIds, err := c.Delete(ctx, &deploy.Ids{Id: id.Id})
	assert.NoError(t, err)
	assert.Equal(t, docId, deletedIds.GetId()[0])

}
func Test_Http(t *testing.T) {

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	payload := new(bytes.Buffer)
	json.NewEncoder(payload).Encode(table2)

	h := GetHandler()
	{
		// add
		c.Request = httptest.NewRequest("POST", "/add", payload)

		h.AddHttp(c)
		b, _ := ioutil.ReadAll(w.Body)
		if w.Code != http.StatusCreated {
			t.Error(w.Code, " ❌ "+string(b))
		}
		id = strings.Trim(string(b), "\"")
		fmt.Println(" ✅ Added", id)
	}
	{
		// get id
		w = httptest.NewRecorder()
		c, _ = gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: id}}

		h.GetHttp(c)
		b, _ := ioutil.ReadAll(w.Body)
		if w.Code != http.StatusOK {
			t.Error(w.Code, " ❌ "+string(b))
		}
		assert.NoError(t, json.Unmarshal(b, &entity))
		fmt.Println(" ✅ Retrieved entity", entity)
	}
	{
		// get all
		c.Request = httptest.NewRequest("GET", "/all", nil)
		h.AllHttp(c)
		b, _ := ioutil.ReadAll(w.Body)
		if w.Code != http.StatusOK {
			t.Error(w.Code, " ❌ "+string(b))
		}
		var entities []database.Entity
		json.Unmarshal(b, &entities)
		fmt.Println(" ✅ Retrieved entities", entities)
		assert.Equal(t, entity, entities[0])
	}
	{
		//  all ids
		c.Request = httptest.NewRequest("GET", "/all", nil)
		h.AllHttpIds(c)
		b, _ := ioutil.ReadAll(w.Body)
		if w.Code != http.StatusOK {
			t.Error(w.Code, " ❌ "+string(b))
		}
		var ids []string
		json.Unmarshal(b, &ids)

		fmt.Println(" ✅ Retrieved ids", ids)
		assert.Equal(t, id, ids[0])
	}
	{
		//  delete
		// c, _ = gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: id}}

		h.DeleteHttp(c)
		b, _ := ioutil.ReadAll(w.Body)
		if w.Code != http.StatusOK {
			t.Error(w.Code, " ❌ "+string(b))
		}
		deletedId := strings.Trim(string(b), "\"")
		assert.Equal(t, id, deletedId)
		fmt.Println(" ✅ Deleted", id)
	}
	{
		// get all
		//c.Request = httptest.NewRequest("GET", "/all", nil)
		h.AllHttp(c)
		b, _ := ioutil.ReadAll(w.Body)
		if w.Code != http.StatusOK {
			t.Error(w.Code, " ❌ "+string(b))
		}
		var entities []database.Entity
		json.Unmarshal(b, &entities)
		assert.Empty(t, entities)
		fmt.Println(" ✅ Is empty")

	}
}

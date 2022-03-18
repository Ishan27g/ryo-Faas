package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	database "github.com/Ishan27g/ryo-Faas/database/db"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var id string
var entity database.Entity
var table1 = map[string]interface{}{
	"Table": "ok",
	"Data": struct {
		Num  int `json:"Num"`
		From int `json:"From"`
		To   int `json:"To"`
	}{
		Num: 1, From: 200, To: 201,
	},
}
var table2 = map[string]interface{}{
	"Table": "ok2",
	"Data": struct {
		Num  int `json:"Num"`
		From int `json:"From"`
		To   int `json:"To"`
	}{
		Num: 1, From: 200, To: 201,
	},
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

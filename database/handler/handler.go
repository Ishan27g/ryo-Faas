package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	database "github.com/Ishan27g/ryo-Faas/database/db"
	deploy "github.com/Ishan27g/ryo-Faas/proto"
	"github.com/Ishan27g/ryo-Faas/types"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

type handle struct {
	Gin *gin.Engine
	database.Database
}

func toDoc(document *deploy.Document) types.NatsDoc {
	var data map[string]interface{}
	err := json.Unmarshal(document.Data, &data)
	if err != nil {
		fmt.Println("toDoc", err.Error())
		return nil
	}
	return types.NewNatsDoc(document.Id, data)
}

func (d handle) forEachDoc(documents *deploy.Documents, cb func(data types.NatsDoc)) (*deploy.Ids, error) {
	var ids *deploy.Ids
	var err error
	for _, document := range documents.Document {
		if doc := toDoc(document); doc != nil {
			cb(doc)
			ids.Id = append(ids.Id, doc.Id())
		} else {
			return nil, err
		}
	}
	return ids, nil
}
func (d handle) New(ctx context.Context, documents *deploy.Documents) (*deploy.Ids, error) {
	return d.forEachDoc(documents, func(doc types.NatsDoc) {
		d.Database.New(doc)
	})
}

func (d handle) Update(ctx context.Context, documents *deploy.Documents) (*deploy.Ids, error) {
	return d.forEachDoc(documents, func(doc types.NatsDoc) {
		d.Database.Update(doc)
	})
}

func (d handle) Get(ctx context.Context, ids *deploy.Ids) (*deploy.Documents, error) {
	var documents []*deploy.Document
	for _, id := range ids.Id {
		if doc := d.Database.Get(id); doc != nil {
			data, _ := json.Marshal((*doc).Document())
			documents = append(documents, &deploy.Document{
				Id:   (*doc).Id(),
				Data: data,
			})
		}
	}
	return &deploy.Documents{Document: documents}, nil
}

func (d handle) Delete(ctx context.Context, ids *deploy.Ids) (*deploy.Ids, error) {
	for _, id := range ids.Id {
		d.Database.Delete(id)
	}
	return nil, nil
}
func (d *handle) AddHttp(c *gin.Context) {
	doc, valid := d.isValid(c)
	if !valid {
		c.JSON(400, nil)
		return
	}
	d.Database.New(doc)
	c.JSON(http.StatusCreated, doc.Id())
}
func (d *handle) UpdateHttp(c *gin.Context) {
	id, found := c.Params.Get("id")
	if !found {
		c.JSON(400, nil)
		return
	}
	doc, valid := d.isValid(c, id)
	if !valid {
		c.JSON(400, nil)
		return
	}
	d.Database.Update(doc)
	c.JSON(http.StatusOK, doc.Document())
}
func (d *handle) DeleteHttp(c *gin.Context) {
	id, found := c.Params.Get("id")
	if !found {
		c.JSON(400, nil)
		return
	}
	d.Database.Delete(id)
	c.JSON(http.StatusOK, id)
}
func (d *handle) GetHttp(c *gin.Context) {
	id, found := c.Params.Get("id")
	if !found {
		c.JSON(400, nil)
		return
	}
	doc := d.Database.Get(id)
	c.JSON(http.StatusOK, (*doc).Document())
}
func (d *handle) AllHttp(c *gin.Context) {
	var docs []map[string]interface{}
	for _, v := range d.Database.All() {
		docs = append(docs, (*v).Document())
	}
	c.JSON(http.StatusOK, docs)
}
func (d *handle) AllHttpIds(c *gin.Context) {
	var docs []string
	for _, v := range d.Database.All() {
		docs = append(docs, (*v).Id())
	}
	c.JSON(http.StatusOK, docs)
}
func (*handle) isValid(c *gin.Context, id ...string) (types.NatsDoc, bool) {
	var data = make(map[string]interface{})
	err := c.ShouldBindJSON(&data)
	if err != nil {
		fmt.Println(err.Error())
		return nil, false
	}
	doc := types.FromJson(data, id...)
	return doc, true
}
func GetHandler() handle {
	h := handle{
		nil,
		database.GetDatabase(),
	}
	gin.SetMode(gin.DebugMode)
	h.Gin = gin.New()
	h.Gin.Use(gin.Recovery())
	h.Gin.Use(otelgin.Middleware("database"))

	g := h.Gin.Group("/database")
	{
		g.GET("/get/:id", h.GetHttp)
		g.GET("/all", h.AllHttp)
		g.GET("/ids", h.AllHttpIds)
		g.POST("/new", h.AddHttp)
		g.PATCH("/update/:id", h.UpdateHttp)
		g.DELETE("/delete/:id", h.DeleteHttp)
	}
	return h

}

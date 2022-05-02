package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	database "github.com/Ishan27g/ryo-Faas/database/db"
	FuncFw "github.com/Ishan27g/ryo-Faas/funcFw"
	deploy "github.com/Ishan27g/ryo-Faas/pkg/proto"
	"github.com/gin-gonic/gin"
)

var db = database.GetDatabase()

type handle struct {
	Gin *gin.Engine
	Rpc rpc
}
type rpc struct{}

func toStoreDoc(document *deploy.Document) database.NatsDoc {
	data := make(map[string]interface{})
	err := json.Unmarshal(document.Data, &data)
	if err != nil {
		fmt.Println("toStoreDoc", err.Error())
		return nil
	}
	return database.NewDocument(document.Table, document.Id, data)
}

func (d *rpc) forEachDoc(documents *deploy.Documents, cb func(data database.NatsDoc)) (*deploy.Ids, error) {
	var ids = new(deploy.Ids)
	var err error
	for _, document := range documents.Document {
		if doc := toStoreDoc(document); doc != nil {
			cb(doc)
			ids.Id = append(ids.Id, doc.Id())
		} else {
			return nil, err
		}
	}
	return ids, nil
}
func (d *rpc) New(ctx context.Context, documents *deploy.Documents) (*deploy.Ids, error) {
	return d.forEachDoc(documents, func(doc database.NatsDoc) {
		db.New(doc)
	})
}

func (d *rpc) Update(ctx context.Context, documents *deploy.Documents) (*deploy.Ids, error) {
	return d.forEachDoc(documents, func(doc database.NatsDoc) {
		db.Update(doc)
	})
}

func (d *rpc) Get(ctx context.Context, ids *deploy.Ids) (*deploy.Documents, error) {
	if len(ids.Id) == 0 {
		return d.All(ctx, ids)
	}
	var documents []*deploy.Document
	for _, id := range ids.Id {
		if entity := db.Get(id); entity != nil {
			data, _ := json.Marshal(entity)
			documents = append(documents, &deploy.Document{
				Id:   entity.Id,
				Data: data,
			})
		}
	}
	return &deploy.Documents{Document: documents}, nil
}

func (d *rpc) Delete(ctx context.Context, ids *deploy.Ids) (*deploy.Ids, error) {
	for _, id := range ids.Id {
		db.Delete(id)
	}
	return &deploy.Ids{Id: ids.Id}, nil
}

func (d *rpc) All(ctx context.Context, ids *deploy.Ids) (*deploy.Documents, error) {
	var documents []*deploy.Document
	for _, entity := range db.All() {
		data, _ := json.Marshal(*entity)
		documents = append(documents, &deploy.Document{
			Id:   entity.Id,
			Data: data,
		})
	}
	return &deploy.Documents{Document: documents}, nil
}
func (d *handle) AddHttp(c *gin.Context) {
	doc, valid := d.isValid(c)
	if !valid {
		c.JSON(400, nil)
		return
	}
	db.New(doc)
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
	fmt.Println(doc)
	db.Update(doc)
	c.JSON(http.StatusOK, doc.Id())
}
func (d *handle) DeleteHttp(c *gin.Context) {
	id, found := c.Params.Get("id")
	if !found {
		c.JSON(400, nil)
		return
	}
	db.Delete(id)
	c.JSON(http.StatusOK, id)
}
func (d *handle) GetHttp(c *gin.Context) {
	id, found := c.Params.Get("id")
	if !found {
		fmt.Println("no id", c.Request.URL)
		c.JSON(400, nil)
		return
	}
	entity := db.Get(id)
	c.JSON(http.StatusOK, *entity)
}
func (d *handle) AfterHttp(c *gin.Context) {
	id, found := c.Params.Get("time")
	if !found {
		c.JSON(400, nil)
		return
	}
	entities := db.After(id)
	var documents []database.Entity
	for _, entity := range entities {
		documents = append(documents, *entity)
	}
	c.JSON(http.StatusOK, entities)
}
func (d *handle) AllHttp(c *gin.Context) {
	var docs []database.Entity
	for _, v := range db.All() {
		docs = append(docs, (*v))
	}
	c.JSON(http.StatusOK, docs)
}
func (d *handle) AllHttpIds(c *gin.Context) {
	var docs []string
	for _, v := range db.All() {
		docs = append(docs, (*v).Id)
	}
	c.JSON(http.StatusOK, docs)
}

type Document struct {
	Table string                 `json:"Table"`
	Data  map[string]interface{} `json:"Data"`
}

func (*handle) isValid(c *gin.Context, id ...string) (database.NatsDoc, bool) {
	var data Document
	err := c.ShouldBindJSON(&data)
	if err != nil {
		fmt.Println(err.Error())
		return nil, false
	}
	doc := database.FromJson(data.Table, data.Data, id...)
	return doc, true
}
func GetHandler() handle {
	h := handle{
		nil,
		rpc{},
	}
	//gin.SetMode(gin.ReleaseMode)
	//h.Gin = gin.New()
	//h.Gin.Use(gin.Recovery())
	//h.Gin.Use(func(ctx *gin.Context) {
	//	fmt.Println(fmt.Sprintf("[%s] [%s]", ctx.Request.Method, ctx.Request.RequestURI))
	//	ctx.Next()
	//})
	//h.Gin.Use(otelgin.Middleware("database"))
	//g := h.Gin.Group("/database")
	//{
	//	g.GET("/get/:id", h.GetHttp)
	//	g.GET("/all", h.AllHttp)
	//	g.GET("/after/:time", h.AfterHttp)
	//	g.GET("/ids", h.AllHttpIds)
	//	g.POST("/new", h.AddHttp)
	//	g.PATCH("/update/:id", h.UpdateHttp)
	//	g.DELETE("/delete/:id", h.DeleteHttp)
	//}

	FuncFw.Export.HttpGin("GetHttp", "/database/get/:id", h.GetHttp)
	FuncFw.Export.HttpGin("AllHttp", "/database/all", h.AllHttp)
	FuncFw.Export.HttpGin("GetHttp", "/database/get/:id", h.GetHttp)
	FuncFw.Export.HttpGin("AfterHttp", "/database/after/:time", h.AfterHttp)
	FuncFw.Export.HttpGin("AddHttp", "/database/new", h.AddHttp)
	FuncFw.Export.HttpGin("UpdateHttp", "/database/update/:id", h.UpdateHttp)
	FuncFw.Export.HttpGin("DeleteHttp", "/database/delete/:id", h.DeleteHttp)
	return h

}

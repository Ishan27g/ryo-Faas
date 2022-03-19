package store

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {

	docStore := Get("payments")

	// data to add
	data := map[string]interface{}{
		"from":   "bob",
		"to":     "alice",
		"amount": 42,
	}

	// add a new `payment` to the db
	id := docStore.Create("", data)
	assert.NotEqual(t, "", id)
	// get it from the db
	dataReturned := docStore.Get(id)

	for _, doc := range (*dataReturned[0]).Document() {
		da := doc.(map[string]interface{})
		m := da["Value"].(map[string]interface{})
		ma := m[id].(map[string]interface{})
		fmt.Println("It is ", m)
		fmt.Println("It is ma", ma)
		//assert.NotNil(t, m[id].(map[string]interface{})["Num"])
		//assert.NotNil(t, m[id].(map[string]interface{})["From"])
		//assert.NotNil(t, m[id].(map[string]interface{})["To"])
	}

	//// dataReturned == data
	//fmt.Println(dataReturned)
	//
	//// update some field
	//data["amount"] = 43
	//docStore.Update(id, data)
	//
	//// delete it
	//docStore.Delete(id)
}

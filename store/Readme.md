# Pub/Sub Document store

Wraps around the `database-client` per document/table name

- `publishes` a nats-message after completing the respective database action

- provides methods to `subscribe`  against database operations

```go
package main

import (
    "github.com/Ishan27g/ryo-Faas/store"
)

func main() {
    // get handler for `payments` document
    docStore := store.Get("payments")

    // data to add
    data := map[string]interface{}{
        "from":   "bob",
        "to":     "alice",
        "amount": 42,
    }

    // subscribe event functions for this document
    go func() {
        go func() {
            docStore.OnCreate(func(document NatsDoc) { // document.Document() == data
                fmt.Println("New payment ")
                document.Print()
            })
        }()
        go func() {
            docStore.OnGet(func(document NatsDoc) {
                fmt.Println("Retrived payment ")
                document.Print()
            })
        }()
        go func() {
            docStore.OnUpdate(func(document NatsDoc) {
                fmt.Println("Updated payment ")
                document.Print()
            })
        }()
        go func() {
            docStore.OnDelete(func(document NatsDoc) {
                fmt.Println("Deleted payment ")
                document.Print()
            })
        }()
    }()

    // add a new `payment` to the db
    id := docStore.Create("", data)

    // get it from the db
    dataReturned := docStore.Get(id)

    // dataReturned == data
    fmt.Println(dataReturned)

    // update some field
    data["amount"] = 43
    docStore.Update(id, data)

    // delete it
    docStore.Delete(id)

}
```

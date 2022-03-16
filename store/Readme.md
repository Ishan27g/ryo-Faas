# Pub/Sub Document store

Acts as a client for the database operations and `publishes` a nats-message after completing the respective database action

- provides methods to `subscribe`  against respective database operations

```go
import "github.com/Ishan27g/ryo-Faas/store"


type DocStore interface {
    // publish

    Create(id string, data map[string]interface{})
    Update(id string, data map[string]interface{})
    Get(id ...string) []*types.NatsDoc
    Delete(id ...string)

    // subscribe

    OnCreate(do EventCb)
    OnUpdate(do EventCb, ids ...string) // subscribe to all ids if nil
    OnDelete(do EventCb, ids ...string) // subscribe to all ids if nil
    OnGet(do EventCb, ids ...string)

    On(subjId string, do EventCb)
}

```

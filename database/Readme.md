# json database over grpc

## client interface

```go
import "github.com/Ishan27g/ryo-Faas/database"

type Client interface {
    New(ctx context.Context, in *deploy.Documents) (*deploy.Ids, error)
    Update(ctx context.Context, in *deploy.Documents) (*deploy.Ids, error)
    Get(ctx context.Context, in *deploy.Ids) (*deploy.Documents, error)
    Delete(ctx context.Context, in *deploy.Ids) (*deploy.Ids, error)
    All(ctx context.Context, in *deploy.Ids) (*deploy.Documents, error)
}
```

- uses <https://github.com/sonyarouje/simdb> as the database

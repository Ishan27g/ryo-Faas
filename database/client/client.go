package database

import (
	"context"
	"fmt"
	"time"

	deploy "github.com/Ishan27g/ryo-Faas/proto"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client interface {
	New(ctx context.Context, in *deploy.Documents) (*deploy.Ids, error)
	Update(ctx context.Context, in *deploy.Documents) (*deploy.Ids, error)
	Get(ctx context.Context, in *deploy.Ids) (*deploy.Documents, error)
	Delete(ctx context.Context, in *deploy.Ids) (*deploy.Ids, error)
	All(ctx context.Context, in *deploy.Ids) (*deploy.Documents, error)
}

func (c *client) New(ctx context.Context, in *deploy.Documents) (*deploy.Ids, error) {
	return c.dc.New(ctx, in)
}

func (c *client) Update(ctx context.Context, in *deploy.Documents) (*deploy.Ids, error) {
	return c.dc.Update(ctx, in)
}

func (c *client) Get(ctx context.Context, in *deploy.Ids) (*deploy.Documents, error) {
	return c.dc.Get(ctx, in)
}

func (c *client) Delete(ctx context.Context, in *deploy.Ids) (*deploy.Ids, error) {
	return c.dc.Delete(ctx, in)
}

type client struct {
	dc deploy.DatabaseClient
}

func (c *client) All(ctx context.Context, in *deploy.Ids) (*deploy.Documents, error) {
	return c.dc.All(ctx, in)
}

// Connect to database, never closed
func Connect(addr string) Client {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	fmt.Println("Connecting to database...", addr)

	grpc.WaitForReady(true)
	grpcClient, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.FailOnNonTempDialError(true),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()))
	if err != nil {
		return nil
	}
	fmt.Println("Connected to database...")
	return &client{dc: deploy.NewDatabaseClient(grpcClient)}
}

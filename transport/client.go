package transport

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	deploy "github.com/Ishan27g/ryo-Faas/proto"
	"github.com/mholt/archiver/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"
)

// AgentWrapper expose an agent with plugins
type AgentWrapper interface {
	Deploy(ctx context.Context, in *deploy.DeployRequest, opts ...grpc.CallOption) (*deploy.DeployResponse, error)
	List(ctx context.Context, in *deploy.Empty, opts ...grpc.CallOption) (*deploy.DeployResponse, error)
	Stop(ctx context.Context, in *deploy.Empty, opts ...grpc.CallOption) (*deploy.DeployResponse, error)
	Details(ctx context.Context, in *deploy.Empty, opts ...grpc.CallOption) (*deploy.DeployResponse, error)
	Upload(ctx context.Context, opts ...grpc.CallOption) (deploy.Deploy_UploadClient, error)
	Logs(ctx context.Context, in *deploy.Function, opts ...grpc.CallOption) (*deploy.Logs, error)
}

type rpcClient struct {
	deploy.DeployClient
	*log.Logger
}

func (r *rpcClient) Deploy(ctx context.Context, in *deploy.DeployRequest, opts ...grpc.CallOption) (*deploy.DeployResponse, error) {
	now := time.Now()
	tr := otel.Tracer("agent-rpc")
	ctxT, span := tr.Start(ctx, "agent-deploy")
	defer span.End()

	rsp, err := r.DeployClient.Deploy(ctxT, in, opts...)

	for _, r := range rsp.Functions {
		span.SetAttributes(attribute.Key("entrypoint").String(r.Entrypoint))
		span.SetAttributes(attribute.Key("status").String(r.Status))
		span.SetAttributes(attribute.Key("url").String(r.Url))
		span.SetAttributes(attribute.Key("time").String(time.Since(now).String()))
	}

	return rsp, err
}

func (r *rpcClient) List(ctx context.Context, in *deploy.Empty, opts ...grpc.CallOption) (*deploy.DeployResponse, error) {
	entrypoint := in.GetEntrypoint()
	if entrypoint == "" {
		return nil, errors.New("no entrypoint in request")
	}
	rsp, err := r.DeployClient.List(ctx, in, opts...)
	return rsp, err
}

func (r *rpcClient) Stop(ctx context.Context, in *deploy.Empty, opts ...grpc.CallOption) (*deploy.DeployResponse, error) {
	rsp, err := r.DeployClient.Stop(ctx, in, opts...)
	return rsp, err
}

func (r *rpcClient) Details(ctx context.Context, in *deploy.Empty, opts ...grpc.CallOption) (*deploy.DeployResponse, error) {
	entrypoint := in.GetEntrypoint()
	if entrypoint == "" {
		return nil, errors.New("no entrypoint in request")
	}
	rsp, err := r.DeployClient.Details(ctx, in, opts...)
	return rsp, err
}

func (r *rpcClient) Upload(ctx context.Context, opts ...grpc.CallOption) (deploy.Deploy_UploadClient, error) {
	return r.DeployClient.Upload(ctx, opts...)
}

func (r *rpcClient) Logs(ctx context.Context, in *deploy.Function, opts ...grpc.CallOption) (*deploy.Logs, error) {
	rsp, err := r.DeployClient.Logs(ctx, in, opts...)
	return rsp, err
}

func ProxyGrpcClient(agentAddr string) AgentWrapper {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	grpc.WaitForReady(true)
	fmt.Println("Connecting to rpc -", agentAddr)
	grpcClient, err := grpc.DialContext(ctx, agentAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	rpc := rpcClient{
		DeployClient: deploy.NewDeployClient(grpcClient),
		Logger:       log.New(os.Stdout, "[RPC-CLIENT] ", log.Ltime),
	}
	return &rpc
}

func compress(dir string) string {
	tmpName := dir + ".zip"
	err := archiver.Archive([]string{dir}, tmpName)
	if err != nil {
		fmt.Println("compress error - ", err.Error())
		return ""
	}
	fmt.Println("compressed ", dir, " to Zip file ", tmpName)
	return tmpName
}
func UploadDir(c deploy.DeployClient, ctx context.Context, dir string, entrypoint string) bool {
	var result = false
	now := time.Now()
	tr := otel.Tracer("transport compress & upload")
	ctx, span := tr.Start(ctx, "agent-upload")
	defer func() {
		span.SetAttributes(attribute.Key("error").Bool(!result))
		span.SetAttributes(attribute.Key("upload").String(time.Since(now).String()))
		span.End()
	}()

	zipFile := compress(dir)

	span.SetAttributes(attribute.Key("compress").String(time.Since(now).String()))
	now = time.Now()

	file, err := os.Open(zipFile)
	if err != nil {
		fmt.Println(err)
		return result
	}
	defer file.Close()
	defer func() {
		os.Remove(zipFile)
	}()

	stream, err := c.Upload(ctx)
	defer stream.CloseAndRecv()

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)
	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("file read error-", err.Error())
			return result
		}
		err = stream.Send(&deploy.File{
			Content:    buffer[:n],
			FileName:   zipFile,
			Entrypoint: entrypoint,
		})
		if err != nil {
			fmt.Println("stream send:", err.Error())
			return result

		}
	}
	result = true
	return result
}

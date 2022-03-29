package transport

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	deploy "github.com/Ishan27g/ryo-Faas/pkg/proto"
	"github.com/mholt/archiver/v3"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// AgentWrapper expose an agent with tracing
type AgentWrapper interface {
	Deploy(ctx context.Context, in *deploy.DeployRequest, opts ...grpc.CallOption) (*deploy.DeployResponse, error)
	Stop(ctx context.Context, in *deploy.Empty, opts ...grpc.CallOption) (*deploy.DeployResponse, error)
	Details(ctx context.Context, in *deploy.Empty, opts ...grpc.CallOption) (*deploy.DeployResponse, error)
	Upload(ctx context.Context, opts ...grpc.CallOption) (deploy.Deploy_UploadClient, error)
}

type rpcClient struct {
	deploy.DeployClient
	*log.Logger
}

func (r *rpcClient) Deploy(ctx context.Context, in *deploy.DeployRequest, opts ...grpc.CallOption) (*deploy.DeployResponse, error) {
	now := time.Now()

	span := trace.SpanFromContext(ctx)

	rsp, err := r.DeployClient.Deploy(ctx, in, opts...)

	for i, r := range rsp.Functions {
		span.AddEvent(printJson(r))
		span.SetAttributes(attribute.Key("entrypoint-" + strconv.Itoa(i)).String(r.Entrypoint))
		span.SetAttributes(attribute.Key("status-" + strconv.Itoa(i)).String(r.Status))
		span.SetAttributes(attribute.Key("url-" + strconv.Itoa(i)).String(r.Url))
		span.SetAttributes(attribute.Key("time-" + strconv.Itoa(i)).String(time.Since(now).String()))
	}

	return rsp, err
}

func (r *rpcClient) Stop(ctx context.Context, in *deploy.Empty, opts ...grpc.CallOption) (*deploy.DeployResponse, error) {
	span := trace.SpanFromContext(ctx)
	rsp, err := r.DeployClient.Stop(ctx, in, opts...)
	span.AddEvent(printJson(r))
	return rsp, err
}

func (r *rpcClient) Details(ctx context.Context, in *deploy.Empty, opts ...grpc.CallOption) (*deploy.DeployResponse, error) {
	span := trace.SpanFromContext(ctx)
	rsp, err := r.DeployClient.Details(ctx, in, opts...)
	span.AddEvent(printJson(rsp))
	return rsp, err
}

func (r *rpcClient) Upload(ctx context.Context, opts ...grpc.CallOption) (deploy.Deploy_UploadClient, error) {
	dc, err := r.DeployClient.Upload(ctx, opts...)
	return dc, err
}

func ProxyGrpcClient(agentAddr string) AgentWrapper {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	grpc.WaitForReady(true)
	grpcClient, err := grpc.DialContext(ctx, agentAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.FailOnNonTempDialError(true),
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()))
	if err != nil {
		// fmt.Println(err.Error())
		return nil
	}
	rpc := rpcClient{
		DeployClient: deploy.NewDeployClient(grpcClient),
		Logger:       log.New(os.Stdout, "[RPC-CLIENT] ", log.Ltime),
	}
	return &rpc
}

func compress(dir string) string {
	//now := time.Now()
	tmpName := dir + ".zip"
	err := archiver.Archive([]string{dir}, tmpName)
	if err != nil {
		fmt.Println("compress error - ", err)
		// if f, e := os.Stat(tmpName); e == nil && f.ModTime().After(now) {
		// 	fmt.Println("compressed ", dir, " to Zip file ", tmpName)
		// 	return tmpName
		// }
		os.RemoveAll(tmpName)
		fmt.Println("deleted - ", tmpName)
		return ""
	}
	fmt.Println("compressed ", dir, " to Zip file ", tmpName)
	return tmpName
}
func UploadDir(c deploy.DeployClient, ctx context.Context, dir string, entrypoint string) bool {
	var result = false
	var zipFile string
	now := time.Now()
	span := trace.SpanFromContext(ctx)
	defer func() {
		span.SetAttributes(attribute.Key("error").Bool(!result))
		span.SetAttributes(attribute.Key("upload").String(time.Since(now).String()))
	}()

	if zipFile = compress(dir); zipFile == "" {
		return result
	}

	span.SetAttributes(attribute.Key("compress").String(time.Since(now).String()))
	now = time.Now()

	file, err := os.Open(zipFile)
	if err != nil {
		span.AddEvent(err.Error())
		fmt.Println(err)
		return result
	}
	defer file.Close()
	defer func() {
		os.Remove(zipFile)
	}()

	ctx = trace.ContextWithSpan(ctx, span)

	stream, err := c.Upload(ctx)
	defer stream.CloseAndRecv()

	span = trace.SpanFromContext(ctx)

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)
	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("file read error-", err.Error())
			span.AddEvent(err.Error())
			return result
		}
		err = stream.Send(&deploy.File{
			Content:    buffer[:n],
			FileName:   zipFile,
			Entrypoint: entrypoint,
		})
		if err != nil {
			fmt.Println("stream send:", err.Error())
			span.AddEvent(err.Error())
			return result

		}
	}
	result = true
	return result
}
func printJson(js interface{}) string {
	data, err := json.MarshalIndent(js, "", " ")
	if err != nil {
		fmt.Println("error:", err)
	}
	return string(data)
}

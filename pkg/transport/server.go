package transport

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	deploy "github.com/Ishan27g/ryo-Faas/pkg/proto"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	"google.golang.org/grpc"
)

type Listener interface {
	Start()
}
type listener struct {
	ctx  context.Context
	grpc *struct {
		deployServer   deploy.DeployServer
		databaseServer deploy.DatabaseServer
		port           string
	}
	*log.Logger
}
type RpcServer struct {
	IsDeploy bool
	Server   interface{}
}
type Config func(*listener)

func WithRpcPort(port string) Config {
	return func(l *listener) {
		l.grpc.port = port
	}
}
func WithDeployServer(d deploy.DeployServer) Config {
	return func(l *listener) {
		l.grpc.deployServer = d
	}
}
func WithDatabaseServer(d deploy.DatabaseServer) Config {
	return func(l *listener) {
		l.grpc.databaseServer = d
	}
}
func Init(ctx context.Context, conf ...Config) Listener {
	l := &listener{}
	l.grpc = &struct {
		deployServer   deploy.DeployServer
		databaseServer deploy.DatabaseServer
		port           string
	}{deployServer: nil, databaseServer: nil, port: ""}
	for _, config := range conf {
		config(l)
	}
	if l.grpc.port == "" {
		fmt.Println("grpc port is nil")
		return nil
	}
	if l.grpc.deployServer == nil && l.grpc.databaseServer == nil {
		fmt.Println("bad grpc")
		return nil
	}
	l.ctx = ctx
	l.Logger = log.New(os.Stdout, "[SERVER]", log.Ltime)
	return l
}

func (l *listener) startGrpc() {
	grpcServer := grpc.NewServer(
		// todo
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)
	if l.grpc.databaseServer != nil {
		deploy.RegisterDatabaseServer(grpcServer, l.grpc.databaseServer)
	} else if l.grpc.deployServer != nil {
		deploy.RegisterDeployServer(grpcServer, l.grpc.deployServer)
	}
	grpcAddr, err := net.Listen("tcp", l.grpc.port)
	if err != nil {
		l.Println(err.Error())
		os.Exit(1)
	}

	go func() {
		l.Println("GRPC server started on " + grpcAddr.Addr().String())
		if err := grpcServer.Serve(grpcAddr); err != nil {
			l.Println("failed to serve: " + err.Error())
			return
		}
	}()
	<-l.ctx.Done()
	grpcServer.Stop()
}

func (l *listener) Start() {
	if l.grpc.port != "" {
		go l.startGrpc()
	}
}

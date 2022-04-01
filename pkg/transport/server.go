package transport

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

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
	http *struct {
		handler http.Handler
		port    string
	}
	*log.Logger
}
type RpcServer struct {
	IsDeploy bool
	Server   interface{}
}
type Config func(*listener)

func WithHandler(handler http.Handler) Config {
	return func(l *listener) {
		l.http.handler = handler
	}
}
func WithHttpPort(port string) Config {
	return func(l *listener) {
		l.http.port = port
	}
}
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
	l.http = &struct {
		handler http.Handler
		port    string
	}{handler: nil, port: ""}
	for _, config := range conf {
		config(l)
	}
	if l.http.port == "" && l.grpc.port == "" {
		fmt.Println("both ports are nil")
		return nil
	}
	if l.grpc.port != "" && l.grpc.deployServer == nil && l.grpc.databaseServer == nil {
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
func (l *listener) startHttp() {

	httpSrv := &http.Server{
		Addr:    l.http.port,
		Handler: l.http.handler,
	}
	go func() {
		l.Println("HTTP started on " + httpSrv.Addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Println("HTTP", err.Error())
		}
	}()
	<-l.ctx.Done()
	cx, can := context.WithTimeout(l.ctx, 2*time.Second)
	defer can()
	if err := httpSrv.Shutdown(cx); err != nil {
		l.Println("Http-Shutdown " + err.Error())
	}
}

func (l *listener) Start() {
	if l.grpc.port != "" {
		go l.startGrpc()
	}
	if l.http.port != "" {
		go l.startHttp()
	}
}

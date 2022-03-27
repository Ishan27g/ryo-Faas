package transport

import (
	"context"
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
	grpc struct {
		server RpcServer
		port   string
	}
	http struct {
		handler http.Handler
		port    string
	}
	*log.Logger
}
type RpcServer struct {
	IsDeploy bool
	Server   interface{}
}

func Init(ctx context.Context, server RpcServer, rpcPort string, handler http.Handler, httpPort string) Listener {
	return &listener{
		ctx: ctx,
		grpc: struct {
			server RpcServer
			port   string
		}{
			server: server,
			port:   rpcPort,
		},
		http: struct {
			handler http.Handler
			port    string
		}{
			handler: handler,
			port:    httpPort,
		},
		Logger: log.New(os.Stdout, "[SERVER]", log.Ltime),
	}
}

func (l *listener) startGrpc() {
	grpcServer := grpc.NewServer(
		// todo
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)
	if l.grpc.server.IsDeploy {
		deploy.RegisterDeployServer(grpcServer, l.grpc.server.Server.(deploy.DeployServer))
	} else {
		deploy.RegisterDatabaseServer(grpcServer, l.grpc.server.Server.(deploy.DatabaseServer))
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

package transport

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	deploy "github.com/Ishan27g/ryo-Faas/proto"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

type Listener interface {
	Start()
}
type listener struct {
	ctx  context.Context
	grpc struct {
		server deploy.DeployServer
		port   string
	}
	http struct {
		engine *gin.Engine
		port   string
	}
	*log.Logger
}

func Init(ctx context.Context, server deploy.DeployServer, rpcPort string, engine *gin.Engine, httpPort string) Listener {
	return &listener{
		ctx: ctx,
		grpc: struct {
			server deploy.DeployServer
			port   string
		}{
			server: server,
			port:   rpcPort,
		},
		http: struct {
			engine *gin.Engine
			port   string
		}{
			engine: engine,
			port:   httpPort,
		},
		Logger: log.New(os.Stdout, "[SERVER ]", log.Ltime),
	}
}

func (l *listener) startGrpc() {
	grpcServer := grpc.NewServer()
	// srv :=
	deploy.RegisterDeployServer(grpcServer, l.grpc.server)
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
		Handler: l.http.engine,
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

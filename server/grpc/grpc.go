package grpc

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"

	"github.com/lukasjarosch/genki/logger"
)

type server struct {
	opts Options
	grpc *grpc.Server
}

func NewServer(opts ...Option) Server {
	options := newOptions(opts...)

	return &server{opts: options, grpc: grpc.NewServer()}
}

// ListenAndServe ties everything together and runs the gRPC server in a separate goroutine.
// The method then blocks until the passed context is cancelled, so this method should also be started
// as goroutine if more work is needed after starting the gRPC server.
func (srv *server) ListenAndServe(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", srv.opts.Port))
	if err != nil {
		logger.Fatalf("failed to listen on tcp port '%s': %s", srv.opts.Port, err.Error())
	}

	go func() {
		logger.Infof("gRPC server running on port %s", srv.opts.Port)
		if err := srv.grpc.Serve(listener); err != nil {
			logger.Errorf("gRPC server crashed: %s", err.Error())
			return
		}
	}()

	<-ctx.Done()
	srv.shutdown()
}

// Server returns the raw grpc Server instance. It is required to register services.
func (srv *server) Server() *grpc.Server {
	return srv.grpc
}

// shutdown is responsible of gracefully shutting down the gRPC server
// First, GracefulStop() is executed. If the call doesn't return
// until the ShutdownGracePeriod is exceeded, the server is terminated directly.
func (srv *server) shutdown() {
	stopped := make(chan struct{})
	go func() {
		srv.grpc.GracefulStop()
		close(stopped)
	}()
	t := time.NewTicker(srv.opts.ShutdownGracePeriod)
	select {
	case <-t.C:
		logger.Warnf("gRPC graceful shutdown timed-out")
	case <-stopped:
		logger.Info("gRPC server shut-down gracefully")
		t.Stop()
	}
}

type Option func(*Options)

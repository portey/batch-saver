package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcLog "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpcRecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/portey/batch-saver/gen/api"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var ErrNotReady = errors.New("grpc service is't ready yet")

func New(port int, service service) (*Server, error) {
	// build service
	srv := Server{
		addr: fmt.Sprintf(":%d", port),
		server: grpc.NewServer(
			grpc.UnaryInterceptor(grpcMiddleware.ChainUnaryServer(
				grpcLog.UnaryServerInterceptor(log.NewEntry(log.StandardLogger())),
				grpcRecovery.UnaryServerInterceptor(),
			)),
			grpc.StreamInterceptor(grpcMiddleware.ChainStreamServer(
				grpcLog.StreamServerInterceptor(log.NewEntry(log.StandardLogger())),
				grpcRecovery.StreamServerInterceptor(),
			)),
		),
	}

	api.RegisterBatchSaverServer(srv.server, newResolver(service))

	return &srv, nil
}

type Server struct {
	addr      string
	server    *grpc.Server
	runErr    error
	readiness bool
}

func (s *Server) Run(ctx context.Context, wg *sync.WaitGroup) {
	log.Info("grpc srv: begin run")
	log.Debug("grpc srv addr:", s.addr)

	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		s.runErr = err
		log.Error("grpc srv: run error: ", err)
		return
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := s.server.Serve(lis)
		log.Error("grpc srv: end run > ", err)
		s.runErr = err
	}()

	go func() {
		<-ctx.Done()
		s.server.GracefulStop()
		log.Info("grpc srv: graceful stop")
	}()

	s.readiness = true
}

func (s *Server) HealthCheck() error {
	if !s.readiness {
		return ErrNotReady
	}
	if s.runErr != nil {
		return fmt.Errorf("grpc service: run issue: %w", s.runErr)
	}
	return nil
}

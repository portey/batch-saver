package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	_ "github.com/lib/pq"
	"github.com/portey/batch-saver/config"
	"github.com/portey/batch-saver/grpc"
	"github.com/portey/batch-saver/healthcheck"
	"github.com/portey/batch-saver/service"
	"github.com/portey/batch-saver/storage/postgres"
	log "github.com/sirupsen/logrus"
)

func main() {
	cfg := config.Read()

	// init logger
	initLogger(cfg.LogLevel)

	log.Info("service starting...")

	// prepare main context
	ctx, cancel := context.WithCancel(context.Background())
	setupGracefulShutdown(cancel)
	var wg = &sync.WaitGroup{}

	storage, err := postgres.New(cfg.PostgresCfg)
	if err != nil {
		log.WithError(err).Fatal("initializing db connection")
	}

	// initializing business logic
	srv := service.New(ctx, cfg.ServiceCfg, storage)

	// initializing grpc server
	grpcSrv, err := grpc.New(cfg.GRPCServerPort, srv)
	if err != nil {
		log.WithError(err).Fatal("tcp server init error")
	}

	// build health check server
	healthSrv := healthcheck.New(cfg.HealthCheckPort, grpcSrv.HealthCheck)

	// run srv
	grpcSrv.Run(ctx, wg)
	healthSrv.Run(ctx, wg)

	// wait while services work
	wg.Wait()
	log.Info("service stopped")
}

func initLogger(logLevel string) {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stderr)

	switch strings.ToLower(logLevel) {
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "trace":
		log.SetLevel(log.TraceLevel)
	default:
		log.SetLevel(log.DebugLevel)
	}
}

func setupGracefulShutdown(stop func()) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChannel
		log.Error("Got Interrupt signal")
		stop()
	}()
}

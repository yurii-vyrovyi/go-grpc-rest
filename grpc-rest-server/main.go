package main

import (
	"context"
	"fmt"
	"github.com/yurii-vyrovyi/grpc-rest-server/internal/grpc"
	"github.com/yurii-vyrovyi/grpc-rest-server/internal/service"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/yurii-vyrovyi/grpc-rest-server/internal/log"
	"github.com/yurii-vyrovyi/grpc-rest-server/internal/opts"

	logger "github.com/sirupsen/logrus"
)

func main() {

	if err := setup(); err != nil {
		logger.Error(err)
		os.Exit(1)
	}
}

func setup() error {

	configFile := os.Getenv("CONFIG_FILE")

	config := opts.Config{}

	err := opts.LoadConfigFromFileOrEnvs(configFile, &config)
	if err != nil {
		return fmt.Errorf("building opts: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	initLogger(config.Log)

	return run(ctx, config)
}

func run(ctx context.Context, config opts.Config) error {

	svc := service.New()

	resolver, err := grpc.NewResolver(svc)
	if err != nil {
		return fmt.Errorf("creating grpc resolver: %w", err)
	}

	grpcServer, err := grpc.NewServer(config.GRPC, resolver)
	if err != nil {
		return fmt.Errorf("creating grpc server: %w", err)
	}

	if err := grpcServer.Run(ctx); err != nil {
		return fmt.Errorf("grpc server run: %w", err)
	}

	return nil
}

func initLogger(config log.Config) {

	var lvl logger.Level

	switch strings.ToUpper(config.Level) {
	case "TRACE":
		lvl = logger.TraceLevel
	case "DEBUG":
		lvl = logger.DebugLevel
	case "INFO":
		lvl = logger.InfoLevel
	case "ERROR":
		lvl = logger.ErrorLevel
	}

	logger.SetLevel(lvl)
	if !config.NonJson {
		logger.SetFormatter(&logger.JSONFormatter{PrettyPrint: config.Pretty})
	}
}

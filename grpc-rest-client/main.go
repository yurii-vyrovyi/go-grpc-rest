package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/yurii-vyrovyi/go-grpc-rest/grpc-rest-client/internal/log"
	"github.com/yurii-vyrovyi/go-grpc-rest/grpc-rest-client/internal/opts"
	grpcV1 "github.com/yurii-vyrovyi/go-grpc-rest/grpc-rest-client/internal/repository/grpc/v1"
	grpcV2 "github.com/yurii-vyrovyi/go-grpc-rest/grpc-rest-client/internal/repository/grpc/v2"
	restV1 "github.com/yurii-vyrovyi/go-grpc-rest/grpc-rest-client/internal/repository/rest/v1"
	restV2 "github.com/yurii-vyrovyi/go-grpc-rest/grpc-rest-client/internal/repository/rest/v2"
	"github.com/yurii-vyrovyi/go-grpc-rest/grpc-rest-client/internal/service"

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

	var sendRepo service.HelloRepo

	switch strings.ToLower(config.Service.Type) {

	case service.TypeGrpcV1:
		grpcRepo, err := grpcV1.BuildRepo(ctx, config.GRPC)
		if err != nil {
			return fmt.Errorf("building hello repo: %w", err)
		}
		defer func() { _ = grpcRepo.Close() }()

		sendRepo = grpcRepo

	case service.TypeGrpcV2:
		grpcRepo, err := grpcV2.BuildRepo(ctx, config.GRPC)
		if err != nil {
			return fmt.Errorf("building hello repo: %w", err)
		}
		defer func() { _ = grpcRepo.Close() }()

		sendRepo = grpcRepo

	case service.TypeRestV1:
		sendRepo = restV1.New(config.Rest)

	case service.TypeRestV2:
		sendRepo = restV2.New(config.Rest)

	default:
		return fmt.Errorf("wrong repository type: %v", config.Service.Type)

	}

	svc := service.New(config.Service, sendRepo)

	if err := svc.Run(ctx); err != nil {
		return fmt.Errorf("service run: %w", err)
	}

	return nil
}

func initLogger(config log.Config) {

	var lvl logger.Level

	switch strings.ToUpper(config.Level) {
	case "TRACE":
		lvl = logger.TraceLevel
	case "INFO":
		lvl = logger.InfoLevel
	case "ERROR":
		lvl = logger.ErrorLevel

	default:
		lvl = logger.DebugLevel
	}

	logger.SetLevel(lvl)
	if !config.NonJson {
		logger.SetFormatter(&logger.JSONFormatter{PrettyPrint: config.Pretty})
	}

}

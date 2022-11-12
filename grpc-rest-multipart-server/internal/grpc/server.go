package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/yurii-vyrovyi/go-grpc-rest/grpc-rest-multipart-server/api"
	_ "github.com/yurii-vyrovyi/go-grpc-rest/grpc-rest-multipart-server/internal/docs"

	"google.golang.org/grpc"
)

type (
	Server struct {
		host           string
		gatewayTimeout time.Duration
		restHost       string
		resolver       *Resolver
	}

	Config struct {
		Host           string        `json:"host" yaml:"host" split_words:"true"`
		GatewayTimeout time.Duration `json:"gatewayTimeout" yaml:"gateway-timeout" split_words:"true"`
		RestHost       string        `json:"restHost" yaml:"rest-host" split_words:"true"`
	}

	SayHelloResponse struct {
		Response string `json:"response"`
	}
)

func NewServer(config Config, resolver *Resolver) (*Server, error) {

	return &Server{
		host:           config.Host,
		gatewayTimeout: config.GatewayTimeout,
		restHost:       config.RestHost,
		resolver:       resolver,
	}, nil
}

func (s *Server) ServeGrpc(grpcServer *grpc.Server) error {

	lis, err := net.Listen("tcp", s.host)
	if err != nil {
		return fmt.Errorf("net.Listen: %w", err)
	}

	return grpcServer.Serve(lis)
}

func (s *Server) Run(ctx context.Context) error {

	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	api.RegisterGrpcRestMultipartServiceServer(grpcServer, s.resolver)

	termChan := make(chan struct{})
	defer close(termChan)

	chanGrpcErr := make(chan error)
	go func() {
		defer close(chanGrpcErr)

		select {
		case chanGrpcErr <- s.ServeGrpc(grpcServer):
		case <-termChan:
		}

	}()

	chanRestErr := make(chan error)
	go func() {
		defer close(chanRestErr)

		select {
		case chanRestErr <- s.ServeHttp(ctx):
		case <-termChan:
		}

	}()

	select {
	case <-ctx.Done():
		return nil

	case err := <-chanGrpcErr:
		return fmt.Errorf("grpc handler: %w", err)

	case err := <-chanRestErr:
		return fmt.Errorf("rest handler: %w", err)
	}

}

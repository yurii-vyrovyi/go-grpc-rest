package grpc

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/yurii-vyrovyi/go-grpc-rest/grpc-rest-server/api"
	_ "github.com/yurii-vyrovyi/go-grpc-rest/grpc-rest-server/internal/docs"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	httpSwagger "github.com/swaggo/http-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

func (s *Server) ServeHttp(ctx context.Context) error {

	ctxDial, cancel := context.WithTimeout(ctx, s.gatewayTimeout)
	defer cancel()

	conn, err := grpc.DialContext(
		ctxDial,
		fmt.Sprintf(s.host),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("dialing grpc connection: %w", err)
	}

	gwmux := runtime.NewServeMux()

	// Register Greeter
	err = api.RegisterGrpcRestServiceHandler(context.Background(), gwmux, conn)
	if err != nil {
		return fmt.Errorf("dialing grpc server: %w", err)
	}

	if err = gwmux.HandlePath("GET", "/swagger/*", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		httpSwagger.WrapHandler(w, r)
	}); err != nil {
		return fmt.Errorf("handle GET /swagger/*: %w", err)
	}

	gwServer := http.Server{
		Addr:              s.restHost,
		Handler:           gwmux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return gwServer.ListenAndServe()
}

func (s *Server) Run(ctx context.Context) error {

	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	api.RegisterGrpcRestServiceServer(grpcServer, s.resolver)

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

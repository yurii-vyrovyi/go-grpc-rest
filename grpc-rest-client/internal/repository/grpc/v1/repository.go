package v1

import (
	"context"
	"errors"
	"fmt"
	"github.com/yurii-vyrovyi/go-grpc-rest/grpc-rest-client/internal/repository/grpc"
	"github.com/yurii-vyrovyi/go-grpc-rest/grpc-rest-client/internal/service"
	"github.com/yurii-vyrovyi/go-grpc-rest/grpc-rest-server/api"

	goGrpc "google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

type (
	Repository struct {
		conn   *goGrpc.ClientConn
		client api.GrpcRestServiceClient
	}
)

func BuildRepo(ctx context.Context, config grpc.Config) (*Repository, error) {

	dialCtx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()

	insecureCreds := insecure.NewCredentials()

	conn, err := goGrpc.DialContext(dialCtx, config.Host, goGrpc.WithTransportCredentials(insecureCreds))
	if err != nil {
		return nil, fmt.Errorf("dialling server: %w", err)
	}

	client := api.NewGrpcRestServiceClient(conn)

	return &Repository{
		conn:   conn,
		client: client,
	}, nil
}

func (r *Repository) Close() error {
	return r.conn.Close()
}

func (r *Repository) Ping() error {
	if r.conn.GetState() != connectivity.Ready {
		return errors.New("gRPC connection is not ready")
	}

	return nil
}

func (r *Repository) SendHello(ctx context.Context, req *service.Request) (string, error) {

	resp, err := r.client.SayHello(ctx, &api.SayHelloRequest{
		Title:       req.Title,
		Description: req.Description,
		IntValue:    int64(req.IntValue),
	})
	if err != nil {
		return "", fmt.Errorf("send hello: %w", err)
	}

	return resp.Response, nil
}

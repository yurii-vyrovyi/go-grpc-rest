package grpc

import (
	"context"

	"github.com/yurii-vyrovyi/go-grpc-rest/grpc-rest-server/api"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service interface {
	ReactOnHello(context.Context, string, string, int, string, []byte) (string, error)
}

type Resolver struct {
	api.UnimplementedGrpcRestServiceServer
	svc Service
}

func NewResolver(svc Service) (*Resolver, error) {
	return &Resolver{
		svc: svc,
	}, nil
}

func (r *Resolver) SayHello(ctx context.Context, req *api.SayHelloRequest) (*api.SayHelloResponse, error) {

	resp, err := r.svc.ReactOnHello(ctx, req.Title, req.Description, int(req.IntValue), req.FileName, req.BinaryData)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.SayHelloResponse{
		Response: resp,
	}, nil
}

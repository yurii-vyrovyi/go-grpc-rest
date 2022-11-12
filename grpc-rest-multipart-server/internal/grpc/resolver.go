package grpc

import (
	"context"

	"github.com/yurii-vyrovyi/go-grpc-rest/grpc-rest-multipart-server/api"
	"github.com/yurii-vyrovyi/go-grpc-rest/grpc-rest-multipart-server/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service interface {
	ReactOnHello(context.Context, string, string, int, []service.Attachment) (string, error)
}

type Resolver struct {
	api.UnimplementedGrpcRestMultipartServiceServer
	svc Service
}

func NewResolver(svc Service) (*Resolver, error) {
	return &Resolver{
		svc: svc,
	}, nil
}

func (r *Resolver) SayHello(ctx context.Context, req *api.SayHelloRequest) (*api.SayHelloResponse, error) {

	attachments := FromApiAttachments(req.Attachments)

	resp, err := r.svc.ReactOnHello(ctx, req.Title, req.Description, int(req.IntValue), attachments)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &api.SayHelloResponse{
		Response: resp,
	}, nil
}

package v2

import (
	"context"
	"fmt"

	"github.com/yurii-vyrovyi/go-grpc-rest/grpc-rest-client/internal/repository/rest"
	"github.com/yurii-vyrovyi/go-grpc-rest/grpc-rest-client/internal/service"
	"github.com/yurii-vyrovyi/go-grpc-rest/grpc-rest-multipart-server/api"
)

type (
	Repository struct {
		rest.Config
	}

	Payload struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		IntValue    int    `json:"int_value"`
	}
)

func New(config rest.Config) *Repository {

	return &Repository{
		Config: config,
	}
}

func (r *Repository) SendHello(ctx context.Context, req *service.Request) (string, error) {

	client := api.NewRestApiClient(r.URL)

	apiAttachments := ToApiAttachments(req.Attachments)

	resp, err := client.SendHello(ctx, &api.SayHelloRequest{
		Title:       req.Title,
		Description: req.Description,
		IntValue:    int64(req.IntValue),
		Attachments: apiAttachments,
	})

	if err != nil {
		return "", fmt.Errorf("sending hello: %w", err)
	}

	return resp.Response, nil
}

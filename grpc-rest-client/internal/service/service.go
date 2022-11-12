package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type (
	HelloRepo interface {
		SendHello(ctx context.Context, req *Request) (string, error)
	}
)

const (
	TypeGrpcV1 = "grpcv1"
	TypeGrpcV2 = "grpcv2"
	TypeRestV1 = "restv1"
	TypeRestV2 = "restv2"
)

type (
	Service struct {
		Config
		helloRepo HelloRepo
	}

	Config struct {
		Type      string   `json:"type" yaml:"type" split_words:"true" validate:"required"`
		DataFiles []string `json:"dataFiles" yaml:"data-files" split_words:"true" validate:"required"`
	}

	Request struct {
		Title       string
		Description string
		IntValue    int
		Attachments []Attachment
	}

	Attachment struct {
		FileName string
		Data     []byte
	}
)

func New(config Config, helloRepo HelloRepo) *Service {

	return &Service{
		Config:    config,
		helloRepo: helloRepo,
	}
}

func (svc *Service) Run(ctx context.Context) error {

	attachments := make([]Attachment, 0, len(svc.DataFiles))
	for _, f := range svc.DataFiles {

		fileName := filepath.Base(f)

		data, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("reading data file: %w", err)
		}

		attachments = append(attachments, Attachment{
			FileName: fileName,
			Data:     data,
		})
	}

	resp, err := svc.helloRepo.SendHello(ctx, &Request{
		Title:       "tit",
		Description: "desc",
		IntValue:    int(time.Now().Unix() % 1000),
		Attachments: attachments,
	})

	if err != nil {
		return fmt.Errorf("sending hello: %w", err)
	}

	fmt.Println("RESPONSE:", resp)

	return nil
}

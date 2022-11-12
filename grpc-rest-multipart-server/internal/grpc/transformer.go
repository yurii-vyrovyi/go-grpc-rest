package grpc

import (
	"github.com/yurii-vyrovyi/go-grpc-rest/grpc-rest-multipart-server/api"
	"github.com/yurii-vyrovyi/go-grpc-rest/grpc-rest-multipart-server/internal/service"
)

func FromApiAttachments(src []*api.Attachment) []service.Attachment {
	if src == nil {
		return nil
	}

	res := make([]service.Attachment, 0, len(src))

	for _, at := range src {
		res = append(res, service.Attachment{
			FileName: at.FileName,
			FileData: at.BinaryData,
		})
	}

	return res
}

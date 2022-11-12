package v2

import (
	"github.com/yurii-vyrovyi/go-grpc-rest/grpc-rest-client/internal/service"
	"github.com/yurii-vyrovyi/go-grpc-rest/grpc-rest-multipart-server/api"
)

func ToApiAttachments(src []service.Attachment) []*api.Attachment {
	if len(src) == 0 {
		return nil
	}

	res := make([]*api.Attachment, 0, len(src))
	for _, at := range src {
		res = append(res, &api.Attachment{
			FileName:   at.FileName,
			BinaryData: at.Data,
		})
	}

	return res
}

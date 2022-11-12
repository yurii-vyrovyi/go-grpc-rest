package grpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/yurii-vyrovyi/go-grpc-rest/grpc-rest-multipart-server/api"

	"github.com/labstack/echo/v4"
	logger "github.com/sirupsen/logrus"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func (s *Server) ServeHttp(_ context.Context) error {
	e := echo.New()

	e.POST("/v2/*", s.V2Handler)
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	return e.Start(s.restHost)
}

func (s *Server) V2Handler(ec echo.Context) error {

	req := ec.Request()

	fmt.Println()
	fmt.Printf("----------- NEW REQUEST (%v) -----------\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Println()

	bodyDump, err := httputil.DumpRequest(req, true)
	if err != nil && !errors.Is(err, io.EOF) {
		logger.Errorf("dumping request: %v", err)
		return ec.NoContent(http.StatusInternalServerError)
	}

	// For small requests it might be interesting to see the whole body
	// fmt.Println(string(bodyDump))

	// For big requests the size will be enough
	fmt.Println("bodyLen:", len(bodyDump))

	fmt.Println()
	fmt.Println()

	mpf, err := ec.MultipartForm()
	if err != nil {
		logger.Errorf("getting multipart form: %v", err)
		return ec.NoContent(http.StatusInternalServerError)
	}

	apiReq := api.SayHelloRequest{}

	// Processing only the first object

	object := mpf.Value["object"]
	if len(object) > 0 {

		if errJson := json.Unmarshal([]byte(object[0]), &apiReq); errJson != nil {
			logger.Errorf("unmarshalling request object: %v", errJson)
			return ec.NoContent(http.StatusInternalServerError)
		}
	}

	// Processing all attachments

	attachments := mpf.File["attachment"]
	apiAttachments := make([]*api.Attachment, 0, len(attachments))

	for _, at := range attachments {

		fileBuf, errRead := readPartFile(attachments[0])
		if errRead != nil {
			logger.Errorf("getting attachment from multiparts: %v", errRead)
			return ec.NoContent(http.StatusInternalServerError)
		}

		apiAttachment := api.Attachment{
			FileName:   at.Filename,
			BinaryData: fileBuf,
		}

		apiAttachments = append(apiAttachments, &apiAttachment)
	}

	apiReq.Attachments = apiAttachments

	resolverResp, err := s.resolver.SayHello(context.Background(), &apiReq)
	if err != nil {
		logger.Errorf("resolver: %v", err)
		return ec.NoContent(http.StatusInternalServerError)
	}

	resp := SayHelloResponse{
		Response: resolverResp.Response,
	}

	ec.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	if err := ec.JSON(http.StatusCreated, resp); err != nil {
		logger.Errorf("creating response: %v", err)
		return ec.NoContent(http.StatusInternalServerError)
	}

	return nil
}

func readPartFile(fh *multipart.FileHeader) ([]byte, error) {

	f, err := fh.Open()
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}

	buf := make([]byte, fh.Size)

	n, err := f.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)

	}

	return buf[:n], nil
}

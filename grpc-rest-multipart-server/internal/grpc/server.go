package grpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/yurii-vyrovyi/go-grpc-rest/grpc-rest-server/api"
	_ "github.com/yurii-vyrovyi/go-grpc-rest/grpc-rest-server/internal/docs"

	logger "github.com/sirupsen/logrus"
	echoSwagger "github.com/swaggo/echo-swagger"
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
	if len(attachments) > 0 {

		at := attachments[0]

		apiReq.FileName = at.Filename

		fileBuf, errRead := readPartFile(attachments[0])
		if errRead != nil {
			logger.Errorf("getting attachment from multiparts: %v", errRead)
			return ec.NoContent(http.StatusInternalServerError)
		}

		apiReq.BinaryData = fileBuf
	}

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

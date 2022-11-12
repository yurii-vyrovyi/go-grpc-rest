package v2

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"
	"time"

	"github.com/yurii-vyrovyi/go-grpc-rest/grpc-rest-client/internal/repository/rest"
	"github.com/yurii-vyrovyi/go-grpc-rest/grpc-rest-client/internal/service"
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

// func (r *Repository) SendHello22(_ context.Context, req *service.Request) error {
//
// 	u, err := url.JoinPath(r.URL, r.SayHelloEndpoint)
// 	if err != nil {
// 		return fmt.Errorf("compiling endpoint [%s] [%s]", r.URL, r.SayHelloEndpoint)
// 	}
//
// 	payload := Payload{
// 		Title:       req.Title,
// 		Description: req.Description,
// 		IntValue:    req.IntValue,
// 	}
//
// 	jsonPayload, err := json.Marshal(payload)
// 	if err != nil {
// 		return fmt.Errorf("marshalling payload: %w", err)
// 	}
//
// 	totalSize := len(jsonPayload) + 4 + 512
// 	for _, at := range req.Attachments {
// 		totalSize += len(at.Data)
// 	}
//
// 	reqBuffer := bytes.NewBuffer(make([]byte, 0, totalSize))
// 	reqBuffer.Write(jsonPayload)
// 	reqBuffer.Write([]byte("\r\n\r\n"))
//
// 	if err := createFileParts(req.Attachments, reqBuffer); err != nil {
// 		return fmt.Errorf("appending files to request: %w", err)
// 	}
//
// 	httpRequest, err := http.NewRequest("POST", u, reqBuffer)
// 	if err != nil {
// 		return fmt.Errorf("creating new request: %w", err)
// 	}
//
// 	httpRequest.Header.Set("Content-Type", "application/json")
//
// 	client := http.Client{Timeout: 5 * time.Second}
//
// 	resp, err := client.Do(httpRequest)
// 	if err != nil {
// 		return fmt.Errorf("sending request: %w", err)
// 	}
//
// 	if resp.StatusCode != http.StatusCreated {
// 		return fmt.Errorf("bad request response: %v [%v]", resp.Status, resp.StatusCode)
// 	}
//
// 	if err := r.handleResponse(resp); err != nil {
// 		return fmt.Errorf("handling response: %w", err)
// 	}
//
// 	return nil
// }

func (r *Repository) SendHello(_ context.Context, req *service.Request) (string, error) {

	u, err := url.JoinPath(r.URL, r.SayHelloEndpoint)
	if err != nil {
		return "", fmt.Errorf("compiling endpoint [%s] [%s]", r.URL, r.SayHelloEndpoint)
	}

	buf := bytes.Buffer{}
	mpw := multipart.NewWriter(&buf)

	if err := createObjectPart(req, mpw); err != nil {
		return "", fmt.Errorf("creating object part: %w", err)
	}

	if err := createFileParts(req.Attachments, mpw); err != nil {
		return "", fmt.Errorf("creating file part: %w", err)
	}

	if err := mpw.Close(); err != nil {
		return "", fmt.Errorf("closing multipart writer: %w", err)
	}

	httpRequest, err := http.NewRequest("POST", u, &buf)
	if err != nil {
		return "", fmt.Errorf("creating new request: %w", err)
	}

	// contentType := fmt.Sprintf("multipart/related; boundary=%s", mpw.Boundary())
	// httpRequest.Header.Set("Content-Type", contentType)

	// httpRequest.Header.Set("Content-Type", "multipart/mixed; boundary="+mpw.Boundary())

	// httpRequest.Header.Set("Content-Length", fmt.Sprintf("%d", buf.Len()))
	httpRequest.Header.Set("Content-Type", mpw.FormDataContentType())

	client := http.Client{Timeout: 5 * time.Second}

	resp, err := client.Do(httpRequest)
	if err != nil {
		return "", fmt.Errorf("sending request: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("bad request response: %v [%v]", resp.Status, resp.StatusCode)
	}

	if err := r.handleResponse(resp); err != nil {
		return "", fmt.Errorf("handling response: %w", err)
	}

	return "", nil
}

func createObjectPart(req *service.Request, mpw *multipart.Writer) error {

	// objWriter, err := mpw.CreateFormField("object")

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"`, "object"))

	h.Set("Content-Type", "application/json")

	objWriter, err := mpw.CreatePart(h)
	if err != nil {
		return fmt.Errorf("creating form field: %w", err)
	}

	payload := Payload{
		Title:       req.Title,
		Description: req.Description,
		IntValue:    req.IntValue,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshalling payload: %w", err)
	}

	_, err = objWriter.Write(jsonPayload)
	if err != nil {
		return fmt.Errorf("writing paload part: %w", err)
	}

	return nil
}

func createFileParts(attachments []service.Attachment, mpw *multipart.Writer) error {

	for _, attachment := range attachments {

		// mediaHeader := textproto.MIMEHeader{}
		// mediaHeader.Set("Content-Type", "application/octet-stream")
		// partWriter, err := mpw.CreatePart(mediaHeader)
		partWriter, err := mpw.CreateFormFile("attachment", attachment.FileName)
		if err != nil {
			return fmt.Errorf("creating form file: %w", err)
		}

		if _, err := io.Copy(partWriter, bytes.NewReader(attachment.Data)); err != nil {
			return fmt.Errorf("copying to part writer: %w", err)
		}

	}

	return nil
}

func (r *Repository) handleResponse(resp *http.Response) error {

	bodyReader := resp.Body
	defer func() { _ = bodyReader.Close() }()

	strLen := resp.Header.Get("Content-Length")
	l, err := strconv.ParseInt(strLen, 10, 64)
	if err != nil {
		l = 4096
	}

	body := make([]byte, l)
	bodyLen, err := bodyReader.Read(body)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("reading body: %w", err)
	}

	fmt.Println("req response:", string(body[:bodyLen]))

	return nil
}

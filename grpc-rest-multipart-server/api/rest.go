package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"time"
)

type RestApiClient struct {
	Host string
}

func NewRestApiClient(host string) *RestApiClient {
	return &RestApiClient{
		Host: host,
	}
}

const (
	EndpointV2SayHello = "/v2/sayhello"
)

func (c *RestApiClient) SendHello(_ context.Context, req *SayHelloRequest) error {

	u, err := url.JoinPath(c.Host, EndpointV2SayHello)
	if err != nil {
		return fmt.Errorf("compiling endpoint [%s] [%s]", c.Host, EndpointV2SayHello)
	}

	buf := bytes.Buffer{}
	mpw := multipart.NewWriter(&buf)

	if errObjectPart := createObjectPart(req, mpw); errObjectPart != nil {
		return fmt.Errorf("creating object part: %w", errObjectPart)
	}

	if errFileParts := createFileParts(req.Attachments, mpw); errFileParts != nil {
		return fmt.Errorf("creating file part: %w", errFileParts)
	}

	if errClose := mpw.Close(); errClose != nil {
		return fmt.Errorf("closing multipart writer: %w", errClose)
	}

	httpRequest, err := http.NewRequest("POST", u, &buf)
	if err != nil {
		return fmt.Errorf("creating new request: %w", err)
	}

	httpRequest.Header.Set("Content-Type", mpw.FormDataContentType())

	client := http.Client{Timeout: 5 * time.Second}

	resp, err := client.Do(httpRequest)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("bad request response: %v [%v]", resp.Status, resp.StatusCode)
	}

	return nil
}

func createObjectPart(req *SayHelloRequest, mpw *multipart.Writer) error {

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"`, "object"))

	h.Set("Content-Type", "application/json")

	objWriter, err := mpw.CreatePart(h)
	if err != nil {
		return fmt.Errorf("creating form field: %w", err)
	}

	// TODO: should be copied in bulk, not field by field
	payload := SayHelloRequest{
		Title:       req.Title,
		Description: req.Description,
		IntValue:    req.IntValue,
		Attachments: nil,
	}

	jsonPayload, err := json.Marshal(&payload)
	if err != nil {
		return fmt.Errorf("marshalling payload: %w", err)
	}

	_, err = objWriter.Write(jsonPayload)
	if err != nil {
		return fmt.Errorf("writing paload part: %w", err)
	}

	return nil
}

func createFileParts(attachments []*Attachment, mpw *multipart.Writer) error {

	for _, attachment := range attachments {

		partWriter, err := mpw.CreateFormFile("attachment", attachment.FileName)
		if err != nil {
			return fmt.Errorf("creating form file: %w", err)
		}

		if _, err := io.Copy(partWriter, bytes.NewReader(attachment.BinaryData)); err != nil {
			return fmt.Errorf("copying to part writer: %w", err)
		}

	}

	return nil
}

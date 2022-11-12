package service

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-multierror"
	logger "github.com/sirupsen/logrus"
	"os"
	"strings"
	"time"
)

type (
	Service struct {
		Config
	}

	Config struct {
		StoreLocation string `json:"storeLocation" yaml:"store-location" split_words:"true" validate:"required"`
	}

	Attachment struct {
		FileName string
		FileData []byte
	}
)

func New(config Config) *Service {
	return &Service{
		Config: config,
	}
}

func (svc *Service) ReactOnHello(
	_ context.Context,
	title, description string, intValue int,
	attachments []Attachment,
) (string, error) {

	logger.Debugf("service got a Hello request: %v", title)

	var resErr error
	savedFiles := make([]string, 0, len(attachments))

	for _, at := range attachments {

		if len(at.FileData) == 0 {
			logger.Infof("%s: [%s: %d]. Data is nil, nothing was saved", title, description, intValue)
			continue
		}

		fullFileName := svc.getFilePathToSave(at.FileName)

		err := os.WriteFile(fullFileName, at.FileData, 0600)
		if err != nil {
			resErr = multierror.Append(resErr, fmt.Errorf("failed to create file: %w", err))
			continue
		}

		savedFiles = append(savedFiles, at.FileName)
	}

	if resErr != nil {
		return "", resErr
	}

	return fmt.Sprintf("%s: [%s: %d]. [%s] were saved",
		title, description, intValue, strings.Join(savedFiles, ",")), nil
}

func (svc *Service) getFilePathToSave(fileName string) string {

	var fn string
	var ext string

	if len(fileName) > 0 {
		fnParts := strings.Split(fileName, ".")
		if len(fnParts) > 0 {
			fn = fnParts[0]
		}

		if len(fnParts) > 1 {
			ext = fnParts[1]
		}
	} else {
		fn = "empty-filename"
		ext = "data"
	}

	fullFilePath := fmt.Sprintf("%s/%s-%s.%s", svc.StoreLocation, fn, time.Now().UTC().Format("2006-01-02T15-04-05"), ext)

	return fullFilePath
}

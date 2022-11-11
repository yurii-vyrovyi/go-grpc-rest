package service

import (
	"context"
	"fmt"
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
)

func New(config Config) *Service {
	return &Service{
		Config: config,
	}
}

func (svc *Service) ReactOnHello(
	_ context.Context,
	title, description string, intValue int,
	fileName string, data []byte,
) (string, error) {

	if data == nil {
		return fmt.Sprintf("%s: [%s: %d]. Data is nil, nothing was saved", title, description, intValue), nil
	}

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
		fn = "empty-fn"
		ext = "data"
	}

	fullFileName := fmt.Sprintf("%s/%s-%s.%s", svc.StoreLocation, fn, time.Now().UTC().Format("2006-01-02T15-04-05"), ext)
	err := os.WriteFile(fullFileName, data, 0600)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}

	logger.Debugf("service got a Hello request: %v", title)

	return fmt.Sprintf("%s: [%s: %d]. [%s] was saved", title, description, intValue, fileName), nil
}

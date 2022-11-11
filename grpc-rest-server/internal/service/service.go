package service

import (
	"context"
	"fmt"
	logger "github.com/sirupsen/logrus"
)

type Service struct {
}

func New() *Service {
	return &Service{}
}

func (svc *Service) ReactOnHello(_ context.Context, title, description string, intValue int) (string, error) {

	logger.Debugf("service got a Hello request: %v", title)

	return fmt.Sprintf("%s: [%s: %d]", title, description, intValue), nil
}

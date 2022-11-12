package grpc

import "time"

type Config struct {
	Host    string        `json:"host" yaml:"host" split_words:"true" validate:"required"`
	Timeout time.Duration `json:"timeout" yaml:"timeout" split_words:"true" default:"15s"`
}

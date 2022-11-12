package rest

type Config struct {
	URL              string `json:"url" yaml:"url" validate:"required"`
	SayHelloEndpoint string `json:"sayHelloEndpoint" yaml:"say-hello-endpoint" validate:"required"`
}

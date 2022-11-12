package opts

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"

	"github.com/yurii-vyrovyi/go-grpc-rest/grpc-rest-multipart-server/internal/grpc"
	"github.com/yurii-vyrovyi/go-grpc-rest/grpc-rest-multipart-server/internal/log"
	"github.com/yurii-vyrovyi/go-grpc-rest/grpc-rest-multipart-server/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/go-multierror"
	"github.com/kelseyhightower/envconfig"
)

const (
	defaultAppPrefix = "APP"
)

type Config struct {
	Log     log.Config     `json:"log" yaml:"log"`
	GRPC    grpc.Config    `json:"grpc" yaml:"grpc"`
	Service service.Config `json:"service" yaml:"service"`
}

type YamlConfigLoader interface {
	Unmarshal([]byte) error
}

func LoadConfigFromFileOrEnvs(configFile string, config interface{}) error {

	if len(configFile) == 0 {
		if err := FromEnv(config); err != nil {
			return fmt.Errorf("locading opts from envs: %w", err)
		}

		return nil
	}

	buf, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("reading opts file: %w", err)
	}

	errYaml := yaml.Unmarshal(buf, config)
	if errYaml == nil {
		return nil
	}

	errJson := json.Unmarshal(buf, config)
	if errJson == nil {
		return nil
	}

	return multierror.Append(
		fmt.Errorf("bad opts file [%s]. Failed to unmarshal YAML and JSON", configFile),
		errYaml, errJson,
	)
}

func FromEnv(config interface{}) error {

	if err := envconfig.Process(defaultAppPrefix, &config); err != nil {
		return fmt.Errorf("failed to load opts: %w", err)
	}

	return nil
}

func (c *Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}

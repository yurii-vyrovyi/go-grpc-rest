package log

type Config struct {
	Level   string `json:"level" yaml:"level" split_words:"true"`
	Pretty  bool   `json:"pretty" yaml:"pretty" split_words:"true"`
	NonJson bool   `json:"nonJson" yaml:"non-json" split_words:"true"`
}

package docs

import (
	"os"

	"github.com/swaggo/swag"
)

type s struct{}

func (s *s) ReadDoc() string {
	doc, err := os.ReadFile("./internal/docs/swagger.yaml")
	if err != nil {
		panic("cannot load swagger.yaml")
	}
	return string(doc)
}

func init() {
	swag.Register(swag.Name, &s{})
}

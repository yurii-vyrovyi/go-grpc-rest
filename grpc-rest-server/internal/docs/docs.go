package docs

import (
	"os"

	"github.com/swaggo/swag"
)

type s struct{}

func (s *s) ReadDoc() string {
	doc, err := os.ReadFile("./api/service.swagger.json")
	if err != nil {
		panic("cannot load protocol.swagger.json")
	}
	return string(doc)
}

func init() {
	swag.Register(swag.Name, &s{})
}

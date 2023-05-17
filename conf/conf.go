package conf

import (
	"bytes"

	toml "github.com/pelletier/go-toml/v2"
)

type Service struct {
	Name  string
	User  string
	Group string
	Git   string
	URLs  map[string]string
	Ports []string
}

type Config struct {
	Services []*Service
}

func Parse(doc []byte) (*Config, error) {
	c := &Config{}
	t := toml.NewDecoder(bytes.NewReader(doc))
	t.DisallowUnknownFields()
	err := t.Decode(c)
	return c, err

}

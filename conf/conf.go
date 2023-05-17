package conf

import (
	"bytes"

	"github.com/miekg/pgo/git"
	toml "github.com/pelletier/go-toml/v2"
)

type Service struct {
	Name       string
	User       string
	Group      string
	Repository string
	URLs       map[string]string
	Ports      []string

	Git *git.Git `toml:"-"`
	// Compose *compose.Compose // podman compose
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

// OSTEMP DIR
// GitTempDir is where we put our git repository each repository get a tmpdir there.
var GitTempDir = "/scratch"

func initializeGitAndCompose(*Service) {
	// mkdir in Git
	// associate name with ssh keys
}

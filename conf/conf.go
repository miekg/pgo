package conf

import (
	"bytes"
	"context"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/miekg/pgo/compose"
	"github.com/miekg/pgo/git"
	toml "github.com/pelletier/go-toml/v2"
	"go.science.ru.nl/log"
)

type Service struct {
	Name       string
	User       string
	Group      string
	Repository string
	Branch     string
	URLs       map[string]string
	Ports      []string
	dir        string // where is repo checked out

	Git     *git.Git         `toml:"-"`
	Compose *compose.Compose `toml:"-"`
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

func (s *Service) InitGitAndCompose() error {
	dir, err := ioutil.TempDir(os.TempDir(), "pgo-*")
	if err != nil {
		return err
	}

	s.Git = git.New(s.Repository, s.User, s.Branch, dir)
	s.Compose = compose.New(s.User, dir)
	return nil
}

// PublicKeys parses the public keys in the ssh/ directory of the repository.
func (s *Service) PublicKeys() ([]ssh.PublicKey, error) {
	if s.dir == "" {
		return nil, fmt.Errorf("local repository path is empty")
	}
	entries, err := os.ReadDir(path.Join(s.dir, "ssh"))
	if err != nil {
		return nil, err
	}
	keys := []ssh.PublicKey{}
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".pub") {
			continue
		}
		pubfile := path.Join(s.dir, "ssh")
		pubfile = path.Join(pubfile, entry.Name())
		log.Infof("Reading public key %q", pubfile)
		data, err := ioutil.ReadFile(pubfile)
		if err != nil {
			continue
		}
		a, _, _, _, err := ssh.ParseAuthorizedKey(data)
		if err != nil {
			continue
		}
		keys = append(keys, a)
	}
	return keys, nil
}

func (s *Service) Track(ctx context.Context, duration time.Duration) {
	log.Infof("Launched tracking routine for %q", s.Name)

	for {
		select {
		case <-time.After(duration):
		}

		if err := s.Git.Checkout(); err != nil {
			log.Warningf("Failed to do initial checkout: %v", err)
			continue
		}

		changed, err := s.Git.Pull([]string{"docker-compose.yml", "docker-compose.yaml"})
		if err != nil {
			log.Warningf("Failed to pull: %v", err)
			continue
		}
		if !changed {
			continue
		}
		s.Compose.Down()
		s.Compose.Build()
		s.Compose.Up()
	}
}

// Track will sha1 sum the contents of file and if it differs from previous runs, will SIGHUP ourselves so we
// exist with status code 2, which in turn will systemd restart us again.
func Track(ctx context.Context, file string, done chan<- os.Signal) {
	hash := ""
	for {
		select {
		case <-time.After(30 * time.Second):
		case <-ctx.Done():
			return
		}
		doc, err := ioutil.ReadFile(file)
		if err != nil {
			log.Warningf("Failed to read config %q: %s", file, err)
			continue
		}
		sha := sha1.New()
		sha.Write(doc)
		hash1 := string(sha.Sum(nil))
		if hash == "" {
			hash = hash1
			continue
		}
		if hash1 != hash {
			log.Info("Config change detected, sending SIGHUP")
			// haste our exit (can this block?)
			done <- syscall.SIGHUP
			return
		}
	}
}

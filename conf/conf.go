package conf

import (
	"bytes"
	"context"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/url"
	"os"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/compose-spec/compose-go/cli"
	"github.com/gliderlabs/ssh"
	"github.com/miekg/pgo/compose"
	"github.com/miekg/pgo/git"
	"github.com/miekg/pgo/osutil"
	toml "github.com/pelletier/go-toml/v2"
	"go.science.ru.nl/log"
)

type Service struct {
	Name        string
	User        string
	Repository  string
	Ignore      bool   // don't restart compose after it got updated if true
	ComposeFile string `toml:"compose,omitempty"` // alternative compose file
	Branch      string
	Import      string            // filename of caddy file to generate
	URLs        map[string]string // url -> host:port
	Env         []string
	Networks    []string
	Git         *git.Git         `toml:"-"`
	Compose     *compose.Compose `toml:"-"`

	dir        string // where is repo checked out
	importdata []byte // caddy's import file data
}

type Config struct {
	Services []*Service
}

func Parse(doc []byte) (*Config, error) {
	c := &Config{}
	t := toml.NewDecoder(bytes.NewReader(doc))
	t.DisallowUnknownFields()
	err := t.Decode(c)
	if err != nil {
		return c, err
	}
	uniq := map[string]struct{}{}
	for _, s := range c.Services {
		if s == nil {
			return c, fmt.Errorf("incomplete service definition")
		}
		if s.Name == "" || s.User == "" || s.Repository == "" {
			return c, fmt.Errorf("expect at least name, user and repository for a service")
		}
		if _, ok := uniq[s.Name]; ok {
			return c, fmt.Errorf("service name %q is not unique", s.Name)
		}
		uniq[s.Name] = struct{}{}

		for u := range s.URLs {
			if _, err := url.Parse(u); err != nil {
				return c, fmt.Errorf("bad url %s for service %q", s.Name, u)
			}
			if _, _, err := net.SplitHostPort(s.URLs[u]); err != nil {
				return c, fmt.Errorf("bad service:port %s for service %q", s.Name, s.URLs[u])
			}
		}
		if s.Branch == "" {
			s.Branch = "main"
		}
		if s.Import != "" {
			s.importdata = MakeCaddyImport(c)
		}
	}

	return c, nil
}

func (s *Service) InitGitAndCompose(dir string) error {
	dir = path.Join(dir, s.Name)
	if err := os.MkdirAll(dir, 0777); err != nil { // all users (possible) in the config, need to access this dir
		return err
	}
	if os.Geteuid() == 0 {
		// chown entire path to correct user
		uid, gid := osutil.User(s.User)
		if err := os.Chown(dir, int(uid), int(gid)); err != nil {
			return err
		}
	}

	s.Git = git.New(s.Repository, s.User, s.Branch, dir)
	s.Compose = compose.New(s.User, dir, s.ComposeFile, s.Networks, s.Env)
	s.dir = dir
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
		data, err := ioutil.ReadFile(pubfile)
		if err != nil {
			continue
		}
		a, _, _, _, err := ssh.ParseAuthorizedKey(data)
		if err != nil {
			log.Warningf("[%s]: Reading public key %q failed: %v", s.Name, pubfile, err)
			continue
		}
		keys = append(keys, a)
	}
	return keys, nil
}

func (s *Service) Track(ctx context.Context, duration time.Duration) {
	log.Infof("[%s]: Launched tracking routine for %q", s.Name, s.Name)

	if err := s.Git.Checkout(); err != nil {
		log.Warningf("[%s]: Failed to do (initial) checkout: %v", s.Name, err)
		return
	}
	var errok error
	if _, err := s.Git.Pull(nil); err != nil {
		log.Warningf("[%s]: Failed to pull: %v", s.Name, err)
		errok = err
	}
	if err := s.Git.Branch(s.Branch); err != nil {
		log.Warningf("[%s]: Failed to checkout branch %s: %v", s.Name, s.Branch, err)
		errok = err
	}
	pubkeys, err := s.PublicKeys()
	if err != nil {
		log.Warningf("[%s]: Failed to get public keys: %v", s.Name, err)
	}

	if errok == nil {
		log.Infof("[%s]: Checked out git repo in %s for %q (branch %s) with %d configured public keys", s.Name, s.dir, s.Name, s.Branch, len(pubkeys))
	} else {
		log.Infof("[%s]: Git repo exist, will fix state in next iteration, last error: %v", s.Name, errok)
	}

	if s.Import != "" && len(s.importdata) > 0 {
		name := path.Join(s.dir, s.Import)
		log.Infof("[%s]: Writing Caddy import file %q", s.Name, s.Import)
		os.WriteFile(name, s.importdata, 0644) // with 644 we shouldn't care about ownership
	}

	if err := s.Compose.AllowedExternalNetworks(); err != nil {
		log.Warningf("[%s]: External network usage outside of allowed networks: %v", s.Name, err)
	} else {
		log.Infof("[%s]: Pulling containers", s.Name)
		if _, err := s.Compose.Pull(nil); err != nil {
			log.Warningf("[%s]: Failed pulling containers: %v", s.Name, err)
		}
		log.Infof("[%s]: Building images", s.Name)
		if _, err := s.Compose.Build(nil); err != nil {
			log.Warningf("[%s]: Failed building containers: %v", s.Name, err)
		}
		log.Infof("[%s]: Upping services", s.Name)
		if _, err := s.Compose.Up(nil); err != nil {
			log.Warningf("[%s]: Failed upping services: %v", s.Name, err)
		}
	}

	namesOfInterest := cli.DefaultFileNames
	if s.ComposeFile != "" {
		namesOfInterest = []string{"s.ComposeFile"}
	}
	for {
		select {
		case <-time.After(jitter(duration)):
		case <-ctx.Done():
			return
		}

		log.Infof("[%s]: Current hash is %q", s.Name, s.Git.Hash())

		changed, err := s.Git.Pull(namesOfInterest)
		if err != nil {
			log.Warningf("[%s]: Failed to pull: %v, deleting repository in %d, and cloning again", s.Name, s.Repository, err)
			if err := s.Git.RemoveAll(); err != nil {
				log.Errorf("[%s]: Failed to remove repository: %v", s.Name, err)
				continue
			}
			if err := s.Git.Checkout(); err != nil {
				log.Warningf("[%s]: Failed to do checkout: %v", s.Name, err)
				continue
			}
			changed = true // force action
		}
		if !changed {
			continue
		}

		if err := s.Compose.AllowedExternalNetworks(); err != nil {
			log.Warningf("[%s]: External network usage outside of allowed networks: %v", s.Name, err)
			continue
		}

		if s.Ignore {
			log.Infof("[%s]: Ignore is set, not restarting any containers", s.Name)
		}

		log.Infof("[%s]: Downing services", s.Name)
		if _, err := s.Compose.Down(nil); err != nil {
			log.Warningf("[%s]: Failed downing services: %v", s.Name, err)
		}
		log.Infof("[%s]: Building images", s.Name)
		if _, err := s.Compose.Build(nil); err != nil {
			log.Warningf("[%s]: Failed building containers: %v", s.Name, err)
		}
		log.Infof("[%s]: Upping services", s.Name)
		if _, err := s.Compose.Up(nil); err != nil {
			log.Warningf("[%s]: Failed upping services: %v", s.Name, err)
		}
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

// jitter will add a random amount of jitter [0, d/2] to d.
func jitter(d time.Duration) time.Duration {
	rand.Seed(time.Now().UnixNano())
	max := d / 2
	return d + time.Duration(rand.Int63n(int64(max)))
}

package conf

import (
	"bytes"
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/gliderlabs/ssh"
	"github.com/miekg/pgo/compose"
	"github.com/miekg/pgo/git"
	"github.com/miekg/pgo/osutil"
	toml "github.com/pelletier/go-toml/v2"
	"go.science.ru.nl/log"
)

const _STOPFILE = ".stop"

type Service struct {
	Name        string
	User        string
	Repository  string
	Registries  []string // user:token@registry auth
	ComposeFile string   `toml:"compose,omitempty"` // alternative compose file
	Branch      string
	Import      string            // filename of caddy file to generate
	Reload      string            // reload command to use for caddy
	Mount       string            // Optional (NFS) mount
	URLs        map[string]string // url -> host:port
	Env         []string
	Networks    []string
	Git         *git.Git         `toml:"-"`
	Compose     *compose.Compose `toml:"-"`

	dir        string   // where is repo checked out
	datadir    string   // where to find the share
	importdata []byte   // caddy's import file data
	reloadcmd  []string // parsed Reload command, should exec service ...
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
		if s.Import != "" && s.Reload == "" {
			// ret error?
			log.Errorf("[%s]: Import is set, but there is no reload command", s.Name)
		}
		if s.Mount != "" && !strings.HasPrefix(s.Mount, "nfs://") {
			return c, fmt.Errorf("bad mount, must start with nfs://")
		}
		if s.Reload != "" {
			reloadcmd, reloadname := "", ""
			_, reloadname, reloadcmd, err = ParseCommand(s.Reload)
			if err != nil {
				return nil, err
			}
			if !strings.HasPrefix(reloadcmd, "exec") {
				return nil, fmt.Errorf("Reload command must start with %q, got %q", "exec", s.reloadcmd)
			}
			// setup reloadcmd to the string after exec, and trim space
			reloadcmd = reloadcmd[len("exec"):]
			reloadcmd = strings.TrimSpace(reloadcmd)

			s.reloadcmd = []string{reloadname}
			s.reloadcmd = append(s.reloadcmd, strings.Fields(reloadcmd)...)
		}
	}

	return c, nil
}

// Stale checks the directory for service subdirs and substracts the current service from it, and then
// downs the compose service and then removes the directory (recursively).
// a slice of stale services that can be downed and removed.
func Stale(sx []*Service, dir string) error {
	ex, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

Stale:
	for _, e := range ex {
		if !e.IsDir() {
			continue
		}
		for i := range sx {
			if sx[i].Name == e.Name() {
				continue Stale
			}
		}

		// TODO(miek): look for /datadir/<service> and umount

		// If the (now deleted) compose config references a non-standard compose, this dance will fail.
		// We _could_ scan for compose variants and pick one... even that would fail, because there can because
		// multiple...
		fulldir := path.Join(dir, e.Name())
		comp := compose.New(e.Name(), "root", fulldir, "", "", nil, nil, nil, "")
		if _, err := comp.Stop(nil); err != nil {
			log.Infof("[%s]: Trying to stop (stale) service %q: %s", e.Name(), e.Name(), err)
		}
		if _, err := comp.Down(nil); err != nil {
			log.Infof("[%s]: Trying to down (stale) service %q: %s", e.Name(), e.Name(), err)
		}
		log.Infof("[%s]: Removing directory: %s", e.Name(), fulldir)
		os.RemoveAll(fulldir)
	}

	return nil
}

func (s *Service) InitGitAndCompose(dir, datadir string) error {
	dir = path.Join(dir, s.Name)
	// TODO(miek) +t here?
	if err := os.MkdirAll(dir, 0777); err != nil { // all users (possible) in the config, need to access this dir
		return err
	}

	datadir = path.Join(datadir, s.Name)
	if err := os.MkdirAll(datadir, 0777); err != nil { // all user need to access the toplevel, dir
		return err
	}
	log.Infof("[%s]: Created git and data dir: %q, %q", s.Name, dir, datadir)
	if os.Geteuid() == 0 {
		// chown last path to correct user
		uid, gid := osutil.User(s.User)
		if err := os.Chown(dir, int(uid), int(gid)); err != nil {
			return err
		}
		if err := os.Chown(datadir, int(uid), int(gid)); err != nil {
			return err
		}
	}

	s.Git = git.New(s.Name, s.Repository, s.User, s.Branch, dir)
	s.Compose = compose.New(s.Name, s.User, dir, s.ComposeFile, datadir, s.Registries, s.Networks, s.Env, s.Mount)
	s.dir = dir
	s.datadir = datadir
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
		data, err := os.ReadFile(pubfile)
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

// Is a <pgodir>/service.stop exists down the service and complain. TODO(miek): read the file's contents for state?
func (s *Service) IsForcedDown() bool {
	stop := s.dir + _STOPFILE
	_, err := os.Stat(stop)
	if err != nil {
		log.Infof("[%s]: Checking stop file %q: %v", s.Name, stop, err)
	} else {
		log.Infof("[%s]: Stop file %q exists", s.Name, stop)
	}
	return !errors.Is(err, os.ErrNotExist)
}

func (s *Service) MountStorage() error {
	if s.Mount == "" {
		return nil
	}
	u, err := url.Parse(s.Mount)
	if err != nil {
		return err
	}
	nfsmount := u.Host + ":" + u.Path
	// TODO(miek): private mounts?
	// TODO(miek): umount -l in Stale?
	args := []string{"-t", "nfs", "-o", "rw,nosuid,hard", nfsmount, s.datadir}
	log.Debugf("Attempting mount %v", args)
	ctx := context.TODO()
	cmd := exec.CommandContext(ctx, "/usr/bin/mount", args...)
	out, err := cmd.CombinedOutput()
	if len(out) > 0 {
		log.Debugf("[%s]: %s", s.Name, string(out))
	}
	return err
}

func (s *Service) Track(ctx context.Context, duration time.Duration) {
	log.Infof("[%s]: Launched tracking routine for %q", s.Name, s.Name)

	err := s.Git.Checkout()
	if err != nil {
		log.Warningf("[%s]: Failed to do check out, will retry: %v", s.Name, err)
		for {
			select {
			case <-time.After(jitter(duration)):
				if err := s.Git.Checkout(); err != nil {
					log.Warningf("[%s]: Failed to do check out, will retry: %v", s.Name, err)
					continue
				}

				break

			case <-ctx.Done():
				return
			}
		}
	}
	log.Infof("[%s]: Succeeded with check out", s.Name)

	if err := s.MountStorage(); err != nil {
		log.Errorf("[%s]: Failed to mount %q: %s", s.Name, s.Mount, err)
	}

	var errok error
	if _, err := s.Git.Pull(nil); err != nil {
		log.Warningf("[%s]: Failed to pull: %v", s.Name, err)
		errok = err
	}
	if err := s.Git.Branch(s.Branch); err != nil {
		log.Warningf("[%s]: Failed to check out branch %s: %v", s.Name, s.Branch, err)
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

	if err := s.Compose.AllowedExternalNetworks(); err != nil {
		log.Warningf("[%s]: External network usage outside of allowed networks: %v", s.Name, err)
	} else if err := s.Compose.AllowedVolumes(); err != nil {
		log.Warningf("[%s]: Volumes' source outside allowed paths: %v", s.Name, err)
	} else if err := s.Compose.Disallow(); err != nil && s.User != "root" {
		log.Errorf("[%s]: Disallowed options used, or generic error: %v", s.Name, err)
	} else {
		log.Infof("[%s]: Pulling containers", s.Name)
		if _, err := s.Compose.Pull(nil); err != nil {
			log.Warningf("[%s]: Failed pulling containers: %v", s.Name, err)
		}
		if s.IsForcedDown() {
			log.Infof("[%s]: Service is forced down, downing to make sure", s.Name)
			if _, err := s.Compose.Down(nil); err != nil {
				log.Warningf("[%s]: Failed downing services: %v", s.Name, err)
			}
		} else {
			log.Infof("[%s]: Upping services", s.Name)
			if _, err := s.Compose.Up(nil); err != nil {
				log.Warningf("[%s]: Failed upping services: %v", s.Name, err)
			}
		}
		log.Infof("[%s]: Tracking upstream from %q", s.Name, s.Git.Hash())
	}

	if s.Import != "" && len(s.importdata) > 0 {
		name := path.Join(s.dir, s.Import)
		log.Infof("[%s]: Writing Caddy import file %q", s.Name, s.Import)
		os.WriteFile(name, s.importdata, 0644) // with 644 we shouldn't care about ownership

		log.Infof("[%s]: Reloading Caddy", s.Name)
		if _, err := s.Compose.Exec(s.reloadcmd); err != nil {
			log.Warningf("[%s]: Failed exec reload command: %v", s.Name, err)
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

		if s.IsForcedDown() {
			log.Infof("[%s]: Service is forced down, downing to make sure", s.Name)
			if _, err := s.Compose.Down(nil); err != nil {
				log.Warningf("[%s]: Failed downing services: %v", s.Name, err)
			}
			continue
		}

		log.Infof("[%s]: Current hash is %q", s.Name, s.Git.Hash())

		changed, err := s.Git.Pull(namesOfInterest)
		if err != nil {
			log.Warningf("[%s]: Failed to pull: %v, deleting repository in %s, and cloning again", s.Name, err, s.Repository)
			if err := s.Git.RemoveAll(); err != nil {
				log.Errorf("[%s]: Failed to remove repository: %v", s.Name, err)
				continue
			}
			if err := s.Git.Checkout(); err != nil {
				log.Warningf("[%s]: Failed to do check out: %v", s.Name, err)
				continue
			}
			changed = true // force action
		}
		if !changed {
			s.Compose.Up(nil) // should be a noop is already running, if not, this hopefully bring the service up
			continue
		}

		if err := s.Compose.AllowedExternalNetworks(); err != nil {
			log.Warningf("[%s]: External network usage outside of allowed networks: %v", s.Name, err)
			continue
		}

		if err := s.Compose.AllowedVolumes(); err != nil {
			log.Warningf("[%s]: Volumes' source outside allowed paths: %v", s.Name, err)
			continue
		}

		if err := s.Compose.Disallow(); err != nil && s.User != "root" {
			log.Errorf("[%s]: Disallowed options used, or generic error: %v", s.Name, err)
			continue
		}

		ex := s.Compose.Extension()
		if !ex.Reload {
			log.Infof("[%s]: reload is set to false, not restarting any containers", s.Name)
		}

		log.Infof("[%s]: Downing services", s.Name)
		if _, err := s.Compose.Down(nil); err != nil {
			log.Warningf("[%s]: Failed downing services: %v", s.Name, err)
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
		doc, err := os.ReadFile(file)
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
	max := d / 2
	return d + time.Duration(rand.Int63n(int64(max)))
}

// ParseCommand parses a pgoctl command line: pgoctl localhost:caddy//exec /bin/ls -l /
// machine: localhost
// name: caddy
// command: exec
// args: rest
func ParseCommand(s string) (machine, name, command string, err error) {
	items := strings.SplitN(s, ":", 2)
	if len(items) != 2 {
		return "", "", "", fmt.Errorf("expected machine:name//command, got %s", s)
	}
	machine = items[0]

	rest := items[1]
	items = strings.Split(rest, "//")
	if len(items) != 2 {
		return "", "", "", fmt.Errorf("expected name//command, got %s", rest)
	}
	name = items[0]
	command = items[1]
	return machine, name, command, nil
}

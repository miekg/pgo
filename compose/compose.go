package compose

import (
	"context"
	"os/exec"
	"sync"

	"github.com/miekg/pgo/metric"
	"github.com/miekg/pgo/osutil"
	"go.science.ru.nl/log"
)

type Compose struct {
	name       string
	user       string   // what user to use
	dir        string   // where to put it
	datadir    string   // datadir from config
	nets       []string // allowed networks from config
	env        []string // extra environment variables
	file       string   // alternate compose file name
	mount      string   // optional mount
	registries []string // private docker registries

	pullLock sync.RWMutex // protects docker pull and hence docker login
}

// New returns a pointer to an intialized Compose.
func New(name, user, directory, file, datadirectory string, registries []string, nets, env []string, mount string) *Compose {
	// TODO(miek): can't use conf.Service here because of import cycle.
	c := &Compose{
		name:       name,
		user:       user,
		dir:        directory,
		datadir:    datadirectory,
		registries: registries,
		nets:       nets,
		file:       file,
		env:        env,
		mount:      mount,
	}
	return c
}

func (c *Compose) run(args ...string) ([]byte, error) {
	ctx := context.TODO()
	if c.file != "" {
		args = append([]string{"--file", c.file}, args...)
	}
	args = append([]string{"compose"}, args...)
	cmd := exec.CommandContext(ctx, "docker", args...)

	if _, err := exec.LookPath("docker-compose"); err == nil {
		// docker-compose is the installed command, use that and strip compose out of args
		args = args[1:]
		cmd = exec.CommandContext(ctx, "docker-compose", args...)
	}

	if err := osutil.RunAs(cmd, c.user); err != nil {
		return nil, err
	}
	cmd.Dir = c.dir
	cmd.Env = append(cmd.Env, c.env...)

	metric.CmdCount.WithLabelValues(c.name, "compose", args[0]).Inc()

	log.Debugf("[%s]: running in %q as %q %v (env: %v)", c.name, cmd.Dir, c.user, cmd.Args, osutil.EnvVars(c.env))

	out, err := cmd.CombinedOutput()
	if len(out) > 0 {
		log.Debugf("[%s]: %s", c.name, string(out))
	}
	if err != nil {
		metric.CmdErrorCount.WithLabelValues(c.name, "compose", args[0]).Inc()
	}

	return out, err
}

func (c *Compose) Down(args []string) ([]byte, error) {
	return c.run(append([]string{"down"}, args...)...)
}
func (c *Compose) Stop(args []string) ([]byte, error) {
	return c.run(append([]string{"stop"}, args...)...)
}
func (c *Compose) Up(args []string) ([]byte, error) {
	// On compose up, I've seen pulls, as well, so take the lock...?
	if len(c.registries) > 0 {
		c.pullLock.Lock()
		defer c.pullLock.Unlock()
	}
	err := c.Login("login")
	if err != nil {
		return nil, err
	}
	log.Infof("[%s]: %s", c.name, "Successfully logged into docker registry")
	if out, err := c.run(append([]string{"up", "-d"}, args...)...); err != nil {
		c.Login("logout")
		return out, err
	}
	err = c.Login("logout")
	return nil, err
}
func (c *Compose) Start(args []string) ([]byte, error) {
	return c.run(append([]string{"start"}, args...)...)
}
func (c *Compose) ReStart(args []string) ([]byte, error) {
	return c.run(append([]string{"start"}, args...)...)
}
func (c *Compose) Pull(args []string) ([]byte, error) {
	// Locking is only needed for private registries.
	if len(c.registries) > 0 {
		c.pullLock.Lock()
		defer c.pullLock.Unlock()
	}
	err := c.Login("login")
	if err != nil {
		return nil, err
	}
	log.Infof("[%s]: %s", c.name, "Successfully logged into docker registry")
	if out, err := c.run(append([]string{"pull"}, args...)...); err != nil {
		c.Login("logout")
		return out, err
	}
	err = c.Login("logout")
	return nil, err
}
func (c *Compose) Logs(args []string) ([]byte, error) {
	return c.run(append([]string{"logs"}, args...)...)
}
func (c *Compose) Ps(args []string) ([]byte, error) {
	return c.run(append([]string{"ps"}, args...)...)
}
func (c *Compose) Exec(args []string) ([]byte, error) {
	return c.run(append([]string{"exec", "-T"}, args...)...)
}

// Load loads the compose files and returns any errors.
func (c *Compose) Load(args []string) ([]byte, error) {
	if err := c.AllowedExternalNetworks(); err != nil {
		return nil, err
	}
	if err := c.AllowedVolumes(); err != nil {
		return nil, err
	}
	if err := c.Disallow(); err != nil {
		return nil, err
	}
	return nil, nil
}

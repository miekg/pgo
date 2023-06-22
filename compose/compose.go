package compose

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/miekg/pgo/metric"
	"github.com/miekg/pgo/osutil"
	"go.science.ru.nl/log"
)

type Compose struct {
	name     string
	user     string   // what user to use
	dir      string   // where to put it
	nets     []string // allowed networks from config
	env      []string // extra environment variables
	file     string   // alternate compose file name
	registry string
}

// New returns a pointer to an intialized Compose.
func New(name, user, directory, file, registry string, nets, env []string) *Compose {
	g := &Compose{
		name:     name,
		user:     user,
		dir:      directory,
		registry: registry,
		nets:     nets,
		file:     file,
		env:      env,
	}
	return g
}

func (c *Compose) run(args ...string) ([]byte, error) {
	ctx := context.TODO()
	if c.file != "" {
		args = append([]string{"-f", c.file}, args...)
	}
	args = append([]string{"compose"}, args...)
	cmd := exec.CommandContext(ctx, "docker", args...)
	if _, err := exec.LookPath("docker-compose"); err == nil {
		// docker-compose is the installed command, use that
		// strip compose out of args
		args = args[1:]
		cmd = exec.CommandContext(ctx, "docker-compose", args...)
	}

	cmd.Dir = c.dir
	path := "/usr/sbin:/usr/bin:/sbin:/bin"
	cmd.Env = []string{env("HOME", osutil.Home(c.user)), env("PATH", path)}

	if os.Geteuid() == 0 {
		uid, gid := osutil.User(c.user)
		if uid == 0 && gid == 0 && c.user != "root" {
			return nil, fmt.Errorf("failed to resolve user %q to uid/gid", c.user)
		}
		dgid := osutil.DockerGroup()
		if dgid == 0 {
			return nil, fmt.Errorf("failed to resolve docker to gid")
		}
		groups := osutil.Groups(c.user)
		groups = append(groups, dgid)
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uid, Gid: gid, Groups: groups}
	}

	envnames := make([]string, len(c.env))
	for i := range c.env {
		cmd.Env = append(cmd.Env, c.env[i])
		fs := strings.Split(c.env[i], "=")
		envnames[i] = fs[0]
	}

	metric.CmdCount.WithLabelValues(c.name, "compose", args[0]).Inc()

	log.Debugf("[%s]: running in %q as %q %v (env: %v)", c.name, cmd.Dir, c.user, cmd.Args, envnames)

	out, err := cmd.CombinedOutput()
	if len(out) > 0 {
		log.Debugf("[%s]: %s", c.name, string(out))
	}
	if err != nil {
		metric.CmdErrorCount.WithLabelValues(c.name, "compose", args[0]).Inc()
	}

	return bytes.TrimSpace(out), err
}

func env(k, v string) string { return k + "=" + v }

func (c *Compose) Build(args []string) ([]byte, error) {
	return c.run(append([]string{"build"}, args...)...)
}
func (c *Compose) Down(args []string) ([]byte, error) {
	return c.run(append([]string{"down"}, args...)...)
}
func (c *Compose) Stop(args []string) ([]byte, error) {
	return c.run(append([]string{"stop"}, args...)...)
}
func (c *Compose) Up(args []string) ([]byte, error) {
	return c.run(append([]string{"up", "-d"}, args...)...)
}
func (c *Compose) Start(args []string) ([]byte, error) {
	return c.run(append([]string{"start"}, args...)...)
}
func (c *Compose) ReStart(args []string) ([]byte, error) {
	return c.run(append([]string{"start"}, args...)...)
}
func (c *Compose) Pull(args []string) ([]byte, error) {
	return c.run(append([]string{"pull"}, args...)...)
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

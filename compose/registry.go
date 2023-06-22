package compose

import (
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

func (c *Compose) Login() error {
	if c.registry == "" {
		return nil
	}

	user := ""
	token := ""
	fields := strings.SplitN(c.registry, ":", 2)
	switch len(fields) {
	case 0:
		println("huh")
	case 1:
		user = c.user
		token = fields[0]
	case 2:
		user = fields[0]
		token = fields[1]
	}

	// do docker login
	// TODO: duplicates compose.go stuff
	ctx := context.TODO()
	log.Infof("[%s]: Performing docker login with %q", c.name, user)
	cmd := exec.CommandContext(ctx, "docker", "login", "-u", user, "-p", token)
	cmd.Dir = c.dir
	path := "/usr/sbin:/usr/bin:/sbin:/bin"
	cmd.Env = []string{env("HOME", osutil.Home(c.user)), env("PATH", path)}

	if os.Geteuid() == 0 {
		uid, gid := osutil.User(c.user)
		if uid == 0 && gid == 0 && c.user != "root" {
			return fmt.Errorf("failed to resolve user %q to uid/gid", c.user)
		}
		dgid := osutil.DockerGroup()
		if dgid == 0 {
			return fmt.Errorf("failed to resolve docker to gid")
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

	metric.CmdCount.WithLabelValues(c.name, "docker", "login").Inc()

	log.Debugf("[%s]: running in %q as %q %v (env: %v)", c.name, cmd.Dir, c.user, cmd.Args[:len(cmd.Args)-1], envnames) // -1 to not leak pw

	out, err := cmd.CombinedOutput()
	if len(out) > 0 {
		log.Debugf("[%s]: %s", c.name, string(out))
	}
	if err != nil {
		metric.CmdErrorCount.WithLabelValues(c.name, "docker", "login").Inc()
	}
	return err
}

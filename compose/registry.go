package compose

import (
	"context"
	"os/exec"
	"strings"

	"github.com/miekg/pgo/metric"
	"github.com/miekg/pgo/osutil"
	"go.science.ru.nl/log"
)

func (c *Compose) Login(login string) error {
	if c.registry == "" {
		return nil
	}

	user := ""
	token := ""
	fields := strings.SplitN(c.registry, ":", 2)
	switch len(fields) {
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
	cmd := exec.CommandContext(ctx, "docker", login, "-u", user, "-p", token)
	if login == "logout" {
		cmd = exec.CommandContext(ctx, "docker", "logout")
	}
	if err := osutil.RunAs(cmd, c.user); err != nil {
		return err
	}
	cmd.Dir = c.dir
	cmd.Env = append(cmd.Env, c.env...)

	metric.CmdCount.WithLabelValues(c.name, "docker", login).Inc()

	if login == "login" {
		log.Debugf("[%s]: running in %q as %q %v (env: %v)", c.name, cmd.Dir, c.user, cmd.Args[:len(cmd.Args)-1], osutil.EnvVars(c.env)) // -1 to not leak pw
	} else {
		log.Debugf("[%s]: running in %q as %q %v (env: %v)", c.name, cmd.Dir, c.user, cmd.Args, osutil.EnvVars(c.env))
	}

	out, err := cmd.CombinedOutput()
	if len(out) > 0 {
		log.Debugf("[%s]: %s", c.name, string(out))
	}
	if err != nil {
		metric.CmdErrorCount.WithLabelValues(c.name, "docker", login).Inc()
	}
	return err
}

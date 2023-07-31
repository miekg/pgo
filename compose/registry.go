package compose

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/miekg/pgo/metric"
	"github.com/miekg/pgo/osutil"
	"go.science.ru.nl/log"
)

func (c *Compose) Login(login string) error {
	if c.registry == nil {
		return nil
	}

	// TODO: locking for this Compose, so other logins don't stump, needs to work in conjunction with docker
	// pull.

	for _, r := range c.registry {
		regi := strings.Index(r, "@")
		if regi < 0 {
			return fmt.Errorf("no @-sign in registry %q", r)
		}
		registry := r[regi+1:]

		user := ""
		token := ""
		fields := strings.SplitN(r[:regi], ":", 2)
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
		log.Infof("[%s]: Performing docker login with %q at %q", c.name, user, registry)
		cmd := exec.CommandContext(ctx, "docker", login, "-u", user, "-p", token, registry)
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
			log.Debugf("[%s]: running in %q as %q %v (env: %v)", c.name, cmd.Dir, c.user, cmd.Args[:len(cmd.Args)-2], osutil.EnvVars(c.env)) // -2 to not leak pw
		} else {
			log.Debugf("[%s]: running in %q as %q %v (env: %v)", c.name, cmd.Dir, c.user, cmd.Args, osutil.EnvVars(c.env))
		}

		out, err := cmd.CombinedOutput()
		if len(out) > 0 {
			log.Debugf("[%s]: %s", c.name, string(out))
		}
		if err != nil {
			metric.CmdErrorCount.WithLabelValues(c.name, "docker", login).Inc()
			return err
		}
	}
	return nil
}

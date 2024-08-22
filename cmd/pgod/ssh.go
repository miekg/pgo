package main

import (
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"

	"github.com/gliderlabs/ssh"
	"github.com/miekg/pgo/conf"
	"github.com/miekg/pgo/metric"
	"github.com/miekg/pgo/osutil"
	"go.science.ru.nl/log"
)

// pgoctl machine dhz//status
func newRouter(c *conf.Config) ssh.Handler {
	return func(ses ssh.Session) {
		pub := ses.PublicKey()
		if pub == nil {
			warnSession(ses, fmt.Sprintf("Connection denied for user %q because no public key supplied", ses.User()), http.StatusUnauthorized)
			return
		}
		if len(ses.Command()) == 0 {
			warnSession(ses, fmt.Sprintf("No commands in connection for user %q", ses.User()), http.StatusBadRequest)
			return
		}
		name, command, args, err := parseCommand(ses.Command())
		if err != nil {
			warnSession(ses, fmt.Sprintf("No correct commands in connection for user %q", ses.User()), http.StatusBadRequest)
			return
		}
		var s *conf.Service
		for i := range c.Services {
			if c.Services[i].Name == name {
				s = c.Services[i]
				break
			}
		}

		if s == nil {
			warnSession(ses, fmt.Sprintf("No service found with name %q", name), http.StatusNotFound)
			return
		}
		// Get the keys and chose *those*
		pubkeys, err := s.PublicKeys()
		if err != nil || len(pubkeys) == 0 {
			warnSession(ses, fmt.Sprintf("No public keys found for %q", name), http.StatusNotFound)
			return
		}

		key := -1
		for i := range pubkeys {
			if ssh.KeysEqual(pubkeys[i], ses.PublicKey()) {
				key = i
				break
			}
		}
		if key == -1 {
			warnSession(ses, fmt.Sprintf("Key for user %q does not match any for name %s", ses.User(), s.Name), http.StatusUnauthorized)
			return
		}

		route, ok := routes[command]
		if !ok {
			warnSession(ses, fmt.Sprintf("Command %q does not match any route", command), http.StatusNotAcceptable)
			return

		}
		log.Infof("[%s]: Routing for user %q, running %q %v", name, ses.User(), command, args)
		out, err := route(s, args)
		exitSession(ses, out, err)
		return
	}
}

var routes = map[string]func(s *conf.Service, args []string) ([]byte, error){
	"up":      func(c *conf.Service, args []string) ([]byte, error) { return c.Compose.Up(args) },
	"down":    func(c *conf.Service, args []string) ([]byte, error) { return c.Compose.Down(args) },
	"stop":    func(c *conf.Service, args []string) ([]byte, error) { return c.Compose.Stop(args) },
	"start":   func(c *conf.Service, args []string) ([]byte, error) { return c.Compose.Start(args) },
	"restart": func(c *conf.Service, args []string) ([]byte, error) { return c.Compose.ReStart(args) },
	"ps":      func(c *conf.Service, args []string) ([]byte, error) { return c.Compose.Ps(args) },
	"pull":    func(c *conf.Service, args []string) ([]byte, error) { return c.Compose.Pull(args) },
	"exec":    func(c *conf.Service, args []string) ([]byte, error) { return c.Compose.Exec(args) },
	"load":    func(c *conf.Service, args []string) ([]byte, error) { return c.Compose.Load(args) },

	"logs": func(c *conf.Service, args []string) ([]byte, error) {
		for i := range args {
			if args[i] == "-f" || args[i] == "--follow" {
				return nil, fmt.Errorf("logs: following logs is not possible")
			}
		}
		return c.Compose.Logs(args)
	},

	"git": func(c *conf.Service, args []string) ([]byte, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("expected git command, got nothing")
		}
		switch args[0] {
		case "pull":
			_, err := c.Git.Pull(nil)
			return nil, err
		case "hash":
			hash := c.Git.Hash()
			return []byte(hash), nil
		default:
			return nil, fmt.Errorf("unrecognized command: %s", args[0])
		}
	},

	"journal": func(c *conf.Service, args []string) ([]byte, error) {
		for i := range args {
			if args[i] == "-f" || args[i] == "--follow" {
				return nil, fmt.Errorf("journal: following logs is not possible")
			}
		}

		uid, _ := osutil.User(c.User)
		args = append([]string{fmt.Sprintf("_UID=%d", uid)}, args...)
		cmd := exec.Command("journalctl", args...)
		log.Debugf("[%s]: running in %q as %q %v", c.Name, cmd.Dir, c.User, cmd.Args)

		out, err := cmd.CombinedOutput()
		if err != nil {
			metric.CmdErrorCount.WithLabelValues(c.Name, "journal", args[0]).Inc()
		}
		return out, err
	},

	"ping": func(c *conf.Service, _ []string) ([]byte, error) {
		return []byte("pong! - " + osutil.Hostname() + "\n"), nil
	},
}

// parseCommand parses: dhz//ps in name (dhz) and command (status) and optional args after it, split on space.
func parseCommand(s []string) (name, command string, args []string, error error) {
	// see conf.go ParseCommand which does more, but also has this code.
	items := strings.Split(s[0], "//")
	if len(items) != 2 {
		return "", "", nil, fmt.Errorf("expected name//command, got %s", s[0])
	}
	name = items[0]
	command = items[1]
	return name, command, s[1:], nil
}

func exitSession(ses ssh.Session, data []byte, err error) {
	if err != nil {
		warnSession(ses, fmt.Sprintf("An error occurred in command in connection for user %q: %s\nCaptured output:\n%s", ses.User(), err, data), http.StatusInternalServerError)
		return
	}
	ses.Write(data)
	ses.Exit(0)
}

func warnSession(ses ssh.Session, warn string, status int) {
	log.Warning(warn)
	io.WriteString(ses, http.StatusText(status)+": "+warn+"\n")
	ses.Exit(status)
}

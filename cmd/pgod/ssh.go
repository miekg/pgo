package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gliderlabs/ssh"
	"github.com/miekg/pgo/conf"
	"go.science.ru.nl/log"
)

// pgoctl machine dhz//status
func newRouter(c *conf.Config) ssh.Handler {
	return func(ses ssh.Session) {
		pub := ses.PublicKey()
		if pub == nil {
			log.Warningf("Connection denied for user %q", ses.User())
			io.WriteString(ses, http.StatusText(http.StatusUnauthorized))
			ses.Exit(http.StatusUnauthorized)
			return
		}
		if len(ses.Command()) == 0 {
			log.Warningf("No commands in connection for user %q", ses.User())
			io.WriteString(ses, http.StatusText(http.StatusBadRequest))
			ses.Exit(http.StatusBadRequest)
			return
		}
		name, command, args, err := parseCommand(ses.Command())
		if err != nil {
			log.Warningf("No correct commands in connection for user %q", ses.User())
			io.WriteString(ses, http.StatusText(http.StatusBadRequest))
			ses.Exit(http.StatusBadRequest)
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
			log.Warningf("No service found with name %q", name)
			io.WriteString(ses, http.StatusText(http.StatusNotFound))
			ses.Exit(http.StatusNotFound)
			return
		}
		// Get the keys and chose *those*
		pubkeys, err := s.PublicKeys()
		if err != nil || len(pubkeys) == 0 {
			log.Warningf("No public keys found for %q", name)
			io.WriteString(ses, http.StatusText(http.StatusNotFound))
			ses.Exit(http.StatusNotFound)
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
			log.Warningf("Key for user %q does not match any for name %s", ses.User(), s.Name)
			io.WriteString(ses, http.StatusText(http.StatusUnauthorized))
			ses.Exit(http.StatusUnauthorized)
			return
		}

		route, ok := routes[command]
		if !ok {
			log.Warningf("Command %q does not match any route", command)
			io.WriteString(ses, http.StatusText(http.StatusNotAcceptable))
			ses.Exit(http.StatusNotAcceptable)
			return

		}
		log.Infof("Routing to %q for user %q", command, ses.User())
		route(s, ses, args)
		return
	}
}

var routes = map[string]func(*conf.Service, ssh.Session, []string){
	"ps": ComposePs,
}

func exitSession(ses ssh.Session, data []byte, err error) {
	if err != nil {
		io.WriteString(ses, http.StatusText(http.StatusInternalServerError))
		ses.Exit(http.StatusInternalServerError)
		return
	}
	ses.Write(data)
	ses.Exit(0)
}

func ComposePs(s *conf.Service, ses ssh.Session, _ []string) {
	out, err := s.Compose.Ps()
	exitSession(ses, out, err)
}

// parseCommand parses: dhz//ps in name (dhz) and command (status) and optional args after it, split on space.
func parseCommand(s []string) (name, command string, args []string, error error) {
	items := strings.Split(s[0], "//")
	if len(items) != 2 {
		return "", "", nil, fmt.Errorf("expected name//command, got %s", s[0])
	}
	name = items[0]
	command = items[1]
	return name, command, s[1:], nil
}

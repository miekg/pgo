package compose

import (
	"fmt"
	"strings"
)

// Disallow parses the compose yaml, and disallows privileged=true, networkmode="host" and ipc="host"
func (c *Compose) Disallow() error {
	comp := Find(c.dir)
	if c.file != "" {
		comp = c.file
	}
	return disallow(comp)
}

// Disallow loads the compose file and checks if any disallowed settings are set.
func disallow(file string) error {
	tp, err := load(file)
	if err != nil {
		return err
	}
	for _, s := range tp.Services {
		if s.Privileged {
			return fmt.Errorf("Service %q sets privileged = true", s.Name)
		}
		if strings.ToLower(s.NetworkMode) == "host" {
			return fmt.Errorf("Service %q sets network_mode = 'host'", s.Name)
		}
		if strings.ToLower(s.Ipc) == "host" {
			return fmt.Errorf("Service %q sets ipc = 'host'", s.Name)
		}
	}
	return nil
}

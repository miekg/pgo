package compose

import (
	"fmt"
	"strings"
)

// Disallow parses the compose yaml, and disallows privileged=true, networkmode="host" and ipc="host"
// It also checks for 'ports' statements as those are _also_ disallowed.
func (c *Compose) Disallow() error {
	comp := Find(c.dir)
	if c.file != "" {
		comp = c.file
	}
	return disallow(comp, c.env)
}

// Disallow loads the compose file and checks if any disallowed settings are set.
func disallow(file string, env []string) error {
	tp, err := load(file, env)
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
		if s.Ports != nil {
			return fmt.Errorf("Service %q uses ports. Use 'expose' instead", s.Name)
		}
	}
	return nil
}

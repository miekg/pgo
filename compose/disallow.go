package compose

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Disallow parses the compose yaml, and disallows privileged=true, networkmode="host" and ipc="host"
// It also checks for 'ports' statements as those are _also_ disallowed.
func (c *Compose) Disallow() error {
	comp := Find(c.dir)
	if c.file != "" {
		comp = filepath.Join(c.dir, c.file)
	}
	return disallow(comp, c.name, c.env)
}

// Disallow loads the compose file and checks if any disallowed settings are set.
func disallow(file, name string, env []string) error {
	tp, err := load(file, name, env)
	if err != nil {
		return err
	}

	if len(tp.Configs) > 0 {
		return fmt.Errorf("Compose file %q uses configs", name)
	}
	if len(tp.Secrets) > 0 {
		return fmt.Errorf("Compose file %q uses secrets", name)
	}

	for _, s := range tp.Services {
		if len(s.SecurityOpt) > 1 {
			// enfore '"no-new-privileges=true"' is set soon.
			return fmt.Errorf("Service %q sets more than 1 security option", s.Name)
		}
		if s.Ipc != "" {
			return fmt.Errorf("Service %q sets ipc", s.Name)
		}
		if s.Privileged {
			return fmt.Errorf("Service %q sets privileged = true", s.Name)
		}
		if len(s.Devices) > 0 {
			return fmt.Errorf("Service %q uses devices", s.Name)
		}
		if len(s.StorageOpt) > 0 {
			return fmt.Errorf("Service %q uses storage opts", s.Name)
		}
		if len(s.CapAdd) > 0 {
			return fmt.Errorf("Service %q want to expand it capabilities", s.Name)
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

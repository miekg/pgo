package compose

import (
	"fmt"
)

// AllowedExternalNetworks returns an error if any of the external networks are not allowed.
func (c *Compose) AllowedExternalNetworks() error {
	if c.nets == nil {
		return nil
	}
	comp := Find(c.dir)
	if c.file != "" {
		comp = c.file
	}
	allnets, err := loadNetworks(comp, c.name, c.env)
	if err != nil {
		return err
	}
	if len(allnets) < 2 {
		return fmt.Errorf("file %q you must have at least 2 (have %v) networks, one of them should be external", comp, allnets)
	}

	nets, err := loadExternalNetworks(comp, c.name, c.env)
	if err != nil {
		return err
	}
	for _, n := range nets {
		ok := false
		for _, n1 := range c.nets {
			if n1 == n {
				ok = true
			}
		}
		if !ok {
			return fmt.Errorf("file %q network %s is not allowed, allowed networks: %v", comp, n, c.nets)
		}
	}

	return nil
}

func loadExternalNetworks(file, name string, env []string) ([]string, error) {
	tp, err := load(file, name, env)
	if err != nil {
		return nil, err
	}
	nets := []string{}
	for _, n := range tp.Networks {
		if n.External {
			nets = append(nets, n.Name) // use n.Name, not n.External.Name
		}
	}
	return nets, nil
}

func loadNetworks(file, name string, env []string) ([]string, error) {
	tp, err := load(file, name, env)
	if err != nil {
		return nil, err
	}
	nets := []string{}
	for _, n := range tp.Networks {
		nets = append(nets, n.Name) // use n.Name, not n.External.Name
	}
	return nets, nil
}

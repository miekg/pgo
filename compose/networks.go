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
	nets, err := loadExternalNetworks(comp)
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
			return fmt.Errorf("network %s is not allowed, allowed networks: %v", n, c.nets)
		}
	}

	return nil
}

// mayby make it have an io.reader instead, of just the yaml?
func loadExternalNetworks(file string) ([]string, error) {
	tp, err := load(file)
	if err != nil {
		return nil, err
	}
	nets := []string{}
	for _, n := range tp.Networks {
		if n.External.External {
			nets = append(nets, n.Name) // use n.Name, not n.External.Name
		}
	}
	return nets, nil
}

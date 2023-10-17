package compose

import (
	"fmt"
	"os"

	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
)

// mayby make it have an io.reader instead, of just the yaml?
func LoadExternalNetworks(file string) ([]string, error) {
	yaml, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	dict, err := loader.ParseYAML([]byte(yaml))
	if err != nil {
		return nil, err
	}
	workingDir, _ := os.Getwd()
	configs := []types.ConfigFile{{Filename: file, Config: dict}}
	config := types.ConfigDetails{
		WorkingDir:  workingDir,
		ConfigFiles: configs,
		Environment: nil,
	}
	tp, err := loader.Load(config)
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

// AllowedExternalNetworks returns an error if any of the external networks are not allowed.
func (c *Compose) AllowedExternalNetworks() error {
	if c.nets == nil {
		return nil
	}
	comp := Find(c.dir)
	if c.file != "" {
		comp = c.file
	}
	nets, err := LoadExternalNetworks(comp)
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

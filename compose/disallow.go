package compose

import (
	"fmt"
	"os"
	"strings"

	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
)

// Disallow parses the compose yaml, and disallows privileged=true, networkmode="host" and ipc="host"
func (c *Compose) Disallow() error {
	comp := Find(c.dir)
	if c.file != "" {
		comp = c.file
	}
	yaml, err := os.ReadFile(comp)
	if err != nil {
		return err
	}
	dict, err := loader.ParseYAML([]byte(yaml))
	if err != nil {
		return err
	}
	workingDir, _ := os.Getwd()
	configs := []types.ConfigFile{{Filename: comp, Config: dict}}
	config := types.ConfigDetails{
		WorkingDir:  workingDir,
		ConfigFiles: configs,
		Environment: nil,
	}
	tp, err := loader.Load(config)
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

package compose

import (
	"fmt"
	"os"
	"strings"

	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
)

// Disallow parse the compose yaml, and disallow privileged=true, networkmode="host" and ipc="host"
func Disallow(file string) error {
	yaml, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	dict, err := loader.ParseYAML([]byte(yaml))
	if err != nil {
		return err
	}
	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	configs := []types.ConfigFile{
		{
			Filename: file,
			Config:   dict,
		},
	}
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
			return fmt.Errorf("Service %q set privileged = true", s.Name)
		}
		if strings.ToLower(s.NetworkMode) == "host" {
			return fmt.Errorf("Service %q set network_mode = 'host'", s.Name)
		}
		if strings.ToLower(s.Ipc) == "host" {
			return fmt.Errorf("Service %q set ipc = 'host'", s.Name)
		}
	}
	return nil
}

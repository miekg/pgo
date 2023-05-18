package compose

import (
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
)

type PortRange struct {
	Lo int
	Hi int
}

func LoadPorts(file string) ([]int, error) {
	yaml, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	dict, err := loader.ParseYAML([]byte(yaml))
	if err != nil {
		return nil, err
	}
	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	configs := []types.ConfigFile{
		{
			Filename: "docker-compose.yml",
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
		return nil, err
	}
	ports := []int{}
	for _, s := range tp.Services {
		for _, p := range s.Ports {
			if p.Published != "" {
				num, err := strconv.ParseUint(p.Published, 10, 32)
				if err != nil {
					continue
				}
				ports = append(ports, int(num))
			}
		}
	}
	return ports, nil
}

// AllowedPorts checks if ports are allowed in the compose file. It returns the first port that is denied, or 0 if
// they are all OK. Any other error is reported via the returned error.
func (c *Compose) AllowedPorts() (int, error) {
	ports, err := LoadPorts(path.Join(c.dir, "docker-compose.yml"))
	if err != nil {
		return 0, err
	}
	for _, p := range ports {
		ok := false
		for _, pr := range c.ports {
			println(p, pr.Lo, pr.Hi)
			if p >= pr.Lo && p <= pr.Hi {
				ok = true
			}
		}
		if !ok {
			return p, fmt.Errorf("port %d is not allowed, allowed ports: %v", p, c.ports)
		}
	}

	return 0, nil
}

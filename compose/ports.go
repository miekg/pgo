package compose

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
)

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
func (c *Compose) AllowedPorts(ports []int) (int, error) {

	return 0, nil
}

// ParsePorts pasrses a n/x string and returns n and n+x, or an error is the string isn't correctly formatted.
func ParsePorts(s string) (int, int, error) {
	items := strings.Split(s, "/")
	if len(items) != 2 {
		return 0, 0, fmt.Errorf("format error, no slash found: %q", s)
	}

	lower, err := strconv.ParseUint(items[0], 10, 32)
	if err != nil {
		return 0, 0, fmt.Errorf("lower ports is not a number: %q", items[0])
	}
	upper, err := strconv.ParseUint(items[1], 10, 32)
	if err != nil {
		return 0, 0, fmt.Errorf("lower ports is not a number: %q", items[0])
	}
	return int(lower), int(lower) + int(upper), nil
}

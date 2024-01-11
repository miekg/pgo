package compose

import (
	"os"

	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
)

func load(file string) (*types.Project, error) {
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
	return loader.Load(config)
}

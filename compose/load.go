package compose

import (
	"context"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
)

func load(file string) (*types.Project, error) {
	o, err := cli.NewProjectOptions([]string{file}) // , opts ...ProjectOptionsFn) (*ProjectOptions, error) {
	if err != nil {
		return nil, err
	}
	//	o.WorkingDir, _ = os.Getwd()
	//	o.ConfigPaths = []string{file}
	// po.Enviroments = mapping
	return cli.ProjectFromOptions(context.TODO(), o)
}

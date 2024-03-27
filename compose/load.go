package compose

import (
	"context"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
)

// TODO(miek): more env stuff, env files??
func load(file, name string, env []string) (*types.Project, error) {
	o, err := cli.NewProjectOptions([]string{file}, cli.WithEnv(env), cli.WithName(name))
	if err != nil {
		return nil, err
	}
	return cli.ProjectFromOptions(context.TODO(), o)
}

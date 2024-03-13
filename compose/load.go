package compose

import (
	"context"
	"os"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
)

// TODO(miek): more env stuff, env files??
func load(file, name string, env []string) (*types.Project, error) {
	wd, _ := os.Getwd()
	o, err := cli.NewProjectOptions([]string{file}, cli.WithEnv(env), cli.WithWorkingDirectory(wd), cli.WithName(name))
	if err != nil {
		return nil, err
	}
	//	o.WorkingDir, _ = os.Getwd()
	return cli.ProjectFromOptions(context.TODO(), o)
}

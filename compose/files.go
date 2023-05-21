package compose

import (
	"os"
	"path/filepath"

	"github.com/compose-spec/compose-go/cli"
)

// Find finds the first compose* file according the default file names of docker-compose.
func Find(pwd string) string {
	for _, n := range cli.DefaultFileNames {
		f := filepath.Join(pwd, n)
		if _, err := os.Stat(f); err == nil {
			return f
		}
	}
	return ""
}

package compose

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/compose-spec/compose-go/v2/types"
)

// AllowedVolumes returns an error if any of the volumes reference a directory outside of c.DataDir
func (c *Compose) AllowedVolumes() error {
	if c.datadir == "" {
		return nil
	}
	comp := Find(c.dir)
	if c.file != "" {
		comp = c.file
	}
	vols, err := loadVolumes(comp, c.name, c.env)
	if err != nil {
		return err
	}
	// These are the Sources for the bind volumes, these must fall either under the pgo dir OR under
	// /<datadir>/<service>. The later is c.dataDir / basename(c.dir)
	datadir := path.Join(c.datadir, path.Base(c.dir))
	for _, v := range vols {
		ok1 := allowedPath(datadir, v)
		ok2 := allowedPath(c.dir, v)
		if !ok1 && !ok2 {
			return fmt.Errorf("volume source path %s does not fall below %q or %q", v, datadir, c.dir)
		}
	}
	return nil
}

func loadVolumes(file, name string, env []string) ([]string, error) {
	tp, err := load(file, name, env)
	if err != nil {
		return nil, err
	}
	vols := []string{}
	for _, s := range tp.Services {
		for _, v := range s.Volumes {
			if v.Type != types.VolumeTypeBind {
				return nil, fmt.Errorf("volumes %s:%s, is not of type %q: %s", v.Source, v.Target, types.VolumeTypeBind, v.Type)
			}
			vols = append(vols, v.Source)
		}
	}
	return vols, nil
}

func allowedPath(base, file string) bool {
	x, err := filepath.Rel(base, file)
	if err != nil {
		return false
	}
	return !strings.Contains(x, "..")
}

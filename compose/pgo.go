package compose

import (
	"fmt"
)

// Extension contains the fields of the PGO extensions in the docker compose file.
type Extension struct {
	Reload bool
}

func (c *Compose) Extension() *Extension {
	comp := Find(c.dir)
	if c.file != "" {
		comp = c.file
	}
	ex, _ := pgo(comp, c.name, c.env)
	return ex
}

// PGO loads the compose file and returns any PGO specific settings, that can be
// added under a top-level: 'x-pgo:'. Currently supported:
//
// - reload: false   # reload/restart all containers if the compose file is updated.
//
// if not set, this defaults to true
func pgo(file, name string, env []string) (*Extension, error) {
	ex := &Extension{Reload: true}
	tp, err := load(file, name, env)
	if err != nil {
		return ex, err
	}
	for k, e := range tp.Extensions {
		if k == "x-pgo" {
			m, ok := e.(map[string]interface{})
			if !ok {
				continue
			}
			for k1, e1 := range m {
				switch k1 {
				case "reload":
					b, ok := e1.(bool)
					if !ok {
						return ex, fmt.Errorf("extension %s is not a boolean: %T", k1, e1)
					}
					ex.Reload = b
				default:
					return ex, fmt.Errorf("unknown extension seen: %s", k1)
				}
			}
		}
	}
	return ex, nil
}

package conf

import (
	"bytes"
	"fmt"
)

func MakeCaddyImport(c *Config) []byte {
	out := &bytes.Buffer{}
	for _, s := range c.Services {
		for u, target := range s.URLs {
			fmt.Fprintf(out, "%s {\n\treverse_proxy %s\n}\n", u, target)
		}
	}
	return out.Bytes()
}

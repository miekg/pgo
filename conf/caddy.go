package conf

import (
	"bytes"
	"fmt"
	"strings"
)

func MakeCaddyImport(c *Config) []byte {
	out := &bytes.Buffer{}
	for _, s := range c.Services {
		for u, target := range s.URLs {
			if !strings.HasPrefix(target, s.Name+"-") {
				continue
			}
			fmt.Fprintf(out, "%s {\n\treverse_proxy %s\n}\n", u, target)
		}
	}
	return out.Bytes()
}

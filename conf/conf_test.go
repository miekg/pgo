package conf

import (
	"fmt"
	"testing"
)

func TestValidConfig(t *testing.T) {
	const conf = `
[[services]]
name = "bliep"
user = "miekg"
group = "miekg"
git = "https://gitlab.science.ru.nl/bla/bliep"
urls = { "slashdot.org" = ":303" }
ports = [ "5005/5", "1025/5" ]
`
	c, err := Parse([]byte(conf))
	if err != nil {
		t.Fatalf("expected to parse config, but got: %s", err)
	}
	for i := range c.Services {
		fmt.Printf("%+v\n", c.Services[i])
	}
}

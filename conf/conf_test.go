package conf

import (
	"testing"
)

func TestValidConfig(t *testing.T) {
	const conf = `
[[services]]
name = "bliep"
user = "miekg"
repository = "https://gitlab.science.ru.nl/bla/bliep"
urls = { "slashdot.org" = "pla:5005" }
`
	_, err := Parse([]byte(conf))
	if err != nil {
		t.Fatalf("expected to parse config, but got: %s", err)
	}
}

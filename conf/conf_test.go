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
repository = "https://gitlab.science.ru.nl/bla/bliep"
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

func TestParsePorts(t *testing.T) {
	a, b, err := parsePorts("1000/5")
	if err != nil {
		t.Fatal(err)
	}
	if a != 1000 || b != 1005 {
		t.Fatalf("expected 1000 and 1005, got %d and %d", a, b)
	}
	a, b, err = parsePorts("1000/0")
	if err != nil {
		t.Fatal(err)
	}
	if a != 1000 || b != 1000 {
		t.Fatalf("expected 1000 and 1000, got %d and %d", a, b)
	}
}

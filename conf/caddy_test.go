package conf

import "testing"

func TestMakeCaddyImportEmpty(t *testing.T) {
	const conf = `
[[services]]
name = "bliep"
user = "miekg"
repository = "https://gitlab.science.ru.nl/bla/bliep"
urls = { "example.org" = "pla:5005" }
`
	c, _ := Parse([]byte(conf))
	out := MakeCaddyImport(c)
	if len(out) != 0 {
		t.Errorf("expected empty caddy import, got length %d", len(out))
	}
}

func TestMakeCaddyImport(t *testing.T) {
	const conf = `
[[services]]
name = "bliep"
user = "miekg"
repository = "https://gitlab.science.ru.nl/bla/bliep"
urls = { "example.org" = "bliep-pla:5005" }
`
	c, _ := Parse([]byte(conf))
	out := MakeCaddyImport(c)
	const expect = `example.org {
	reverse_proxy bliep-pla:5005
}
`
	if expect != string(out) {
		t.Errorf("generated output doesn't match expected")
		t.Logf("expect = %s\ngot = %s\n", expect, string(out))
	}
}

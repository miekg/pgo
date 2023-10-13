package conf

import "testing"

func TestMakeCaddyImport(t *testing.T) {
	const conf = `
[[services]]
name = "bliep"
user = "miekg"
repository = "https://gitlab.science.ru.nl/bla/bliep"
urls = { "example.org" = "pla:5005" }
`
	c, _ := Parse([]byte(conf))
	out := MakeCaddyImport(c)
	const expect = `example.org {
	reverse_proxy pla:5005
}
`
	if expect != string(out) {
		t.Errorf("generated output doesn't match expected")
		t.Logf("expect = %s\ngot = %s\n", expect, string(out))
	}
}

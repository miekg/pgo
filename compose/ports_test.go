package compose

import (
	"testing"
)

func TestLoadPorts(t *testing.T) {
	ports, err := LoadPorts("testdata/docker-compose.yml")
	if err != nil {
		t.Fatal(err)
	}
	if ports[0] != 8080 {
		t.Fatalf("expected port 8080, got %d", ports[0])
	}
}

func TestParsePorts(t *testing.T) {
	a, b, err := ParsePorts("1000/5")
	if err != nil {
		t.Fatal(err)
	}
	if a != 1000 || b != 1005 {
		t.Fatalf("expected 1000 and 1005, got %d and %d", a, b)
	}
}

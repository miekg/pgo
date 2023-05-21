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

func TestAllowedPorts(t *testing.T) {
	c := &Compose{
		user:  "",
		dir:   "testdata",
		ports: []PortRange{{1000, 1005}},
	}
	// disallowed
	port, err := c.AllowedPorts()
	if err == nil {
		t.Fatal("expected error, got none")
	}
	if port != 8080 {
		t.Fatalf("expected port 8080, got %d: %s", port, err)
	}
	// allowed
	c.ports = []PortRange{{8080, 8081}, {9090, 9091}}
	port, err = c.AllowedPorts()
	if err != nil {
		t.Fatal(err)
	}
	if port != 0 {
		t.Fatalf("expected port 0 (=ok), got %d", port)
	}
}

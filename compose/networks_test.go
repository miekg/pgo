package compose

import (
	"testing"
)

func TestLoadExternalNetworks(t *testing.T) {
	nets, err := loadExternalNetworks("testdata/docker-compose.yml")
	if err != nil {
		t.Fatal(err)
	}
	if nets[0] != "reverse_proxy" {
		t.Fatalf("expected network 'reverse_proxy', got %s", nets[0])
	}
}

func TestAllowedExternalNetworks(t *testing.T) {
	c := &Compose{
		user: "",
		dir:  "testdata",
		nets: []string{"a", "b"},
	}
	// disallowed
	err := c.AllowedExternalNetworks()
	if err == nil {
		t.Fatal("expected error, got none")
	}
	// allowed
	c.nets = []string{"a", "b", "reverse_proxy"}
	err = c.AllowedExternalNetworks()
	if err != nil {
		t.Fatal(err)
	}
}

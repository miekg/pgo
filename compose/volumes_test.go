package compose

import (
	"os"
	"testing"
)

func TestLoadVolumes(t *testing.T) {
	_, err := loadVolumes("testdata/docker-compose_volumes.yml", "", nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAllowedVolumes(t *testing.T) {
	wd, _ := os.Getwd()
	c := &Compose{
		user:    "",
		dir:     wd,
		file:    "testdata/docker-compose_volumes.yml",
		datadir: "/data",
	}
	// allowed
	if err := c.AllowedVolumes(); err != nil {
		t.Fatalf("expected no error, got: %s", err)
	}
	c.dir = "/tmp"
	if err := c.AllowedVolumes(); err == nil {
		t.Fatal("expected an error, got none")
	}
}

package compose

import "testing"

func TestDisallow(t *testing.T) {
	err := disallow("testdata/docker-compose_priv.yml", "", nil)
	if err == nil {
		t.Fatal("expected error, got none")
	}
	t.Logf(err.Error())
}

func TestDisallowPorts(t *testing.T) {
	err := disallow("testdata/docker-compose_ports.yml", "", nil)
	if err == nil {
		t.Fatal("expected error, got none")
	}
	t.Logf(err.Error())
}

func TestDisallowConfigs(t *testing.T) {
	err := disallow("testdata/docker-compose_configs.yml", "", nil)
	if err == nil {
		t.Fatal("expected error, got none")
	}
	t.Logf(err.Error())
}

func TestDisallowIncludes(t *testing.T) {
	err := disallow("testdata/docker-compose_include.yml", "", nil)
	if err == nil {
		t.Fatal("expected error, got none")
	}
	t.Logf(err.Error())
}

func TestDisallowPrivileged(t *testing.T) {
	err := disallow("testdata/docker-compose_privileged.yml", "", nil)
	if err == nil {
		t.Fatal("expected error, got none")
	}
	t.Logf(err.Error())
}

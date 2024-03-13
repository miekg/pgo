package compose

import (
	"testing"
)

func TestPgo(t *testing.T) {
	ex, err := pgo("testdata/x-docker-compose.yml", nil)
	if err != nil {
		t.Fatal(err)
	}
	if ex.Reload {
		t.Errorf("expected reload to be false, got true")
	}
}

func TestPgoEnvironment(t *testing.T) {
	_, err := pgo("testdata/docker-compose-env.yml", []string{"APP_PORT=100"})
	if err != nil {
		t.Fatal(err)
	}
}

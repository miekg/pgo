package compose

import "testing"

func TestDisallow(t *testing.T) {
	err := disallow("testdata/docker-compose_priv.yml")
	if err == nil {
		t.Fatal("expected error, got none")
	}
	t.Logf(err.Error())
}

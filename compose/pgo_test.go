package compose

import (
	"testing"
)

func TestPgo(t *testing.T) {
	ex, err := pgo("testdata/x-docker-compose.yml")
	if err != nil {
		t.Fatal(err)
	}
	if ex.reload {
		t.Errorf("expected reload to be false, got true")
	}
}

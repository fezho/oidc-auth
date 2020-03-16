package memory

import (
	"github.com/fezho/oidc-auth-service/storage/testutils"
	"testing"
)

func TestMemoryStorage(t *testing.T) {
	cfg := &Config{SessionConfig: testutils.MockSessionConfig()}
	s, err := cfg.Open()
	if err != nil {
		t.Fatal("failed to open memory storage", err)
	}
	defer s.Close()

	testutils.RunTestNew(t, s)
	testutils.RunTestGet(t, s)
	testutils.RunTestSave(t, s)
	testutils.RunTestMaxAge(t, s)
}

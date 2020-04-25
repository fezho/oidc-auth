package memory_test

import (
	"testing"

	"github.com/fezho/oidc-auth/storage/memory"
	"github.com/fezho/oidc-auth/storage/testutils"
)

func TestMemoryStorage(t *testing.T) {
	cfg := &memory.Config{SessionConfig: testutils.MockSessionConfig()}
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

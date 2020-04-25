package bolt_test

import (
	"os"
	"testing"

	"github.com/fezho/oidc-auth/storage/bolt"
	"github.com/fezho/oidc-auth/storage/testutils"
)

func TestBoltStorage(t *testing.T) {
	path := "/tmp/data.db"
	cfg := &bolt.Config{
		Path:          path,
		BucketName:    "session",
		SessionConfig: testutils.MockSessionConfig(),
	}
	s, err := cfg.Open()
	if err != nil {
		t.Fatal("failed to open redis storage", err)
	}
	defer func() {
		_ = s.Close()       // nolint
		_ = os.Remove(path) // nolint
	}()

	testutils.RunTestNew(t, s)
	testutils.RunTestGet(t, s)
	testutils.RunTestSave(t, s)
	testutils.RunTestMaxAge(t, s)
}

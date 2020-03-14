package bolt

import (
	"github.com/fezho/oidc-auth-service/storage/testutils"
	"os"
	"testing"
)

func TestBoltStorage(t *testing.T) {
	path := "/tmp/data.db"
	cfg := &Config{
		Path:          path,
		BucketName:    "session",
		SessionConfig: testutils.MockSessionConfig("gob"),
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

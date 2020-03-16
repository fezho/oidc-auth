package redis

import (
	"github.com/fezho/oidc-auth-service/storage/testutils"
	"testing"
)

func TestRedisStorage(t *testing.T) {
	cfg := &Config{
		Address: "127.0.0.1:6379",
		//Password:      "",
		KeyPrefix:     "session",
		SessionConfig: testutils.MockSessionConfig(),
	}
	s, err := cfg.Open()
	if err != nil {
		t.Fatal("failed to open redis storage", err)
	}
	defer s.Close()

	testutils.RunTestNew(t, s)
	testutils.RunTestGet(t, s)
	testutils.RunTestSave(t, s)
	testutils.RunTestMaxAge(t, s)
}

package internal_test

import (
	"testing"

	"github.com/gorilla/sessions"

	"github.com/fezho/oidc-auth/storage/internal"
)

func TestEncode_Decode(t *testing.T) {
	ss1 := sessions.NewSession(nil, "hello")
	ss1.Values["key"] = "value"

	b, err := internal.Encode(ss1)
	if err != nil {
		t.Fatal("failed to encode ss1", err)
	}

	if len(b) == 0 {
		t.Fatal("expected to not empty bytes, got nil")
	}

	ss2 := sessions.NewSession(nil, "hello")
	if err := internal.Decode(b, ss2); err != nil {
		t.Fatal("failed to decode session from data", err)
	}

	value, ok := ss2.Values["key"]
	if !ok || value != "value" {
		t.Fatal("expected to get key-value pair, got wrong")
	}
}

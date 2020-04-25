package server_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/fezho/oidc-auth/server"
	"github.com/fezho/oidc-auth/storage/memory"
)

var (
	_httpServer *httptest.Server
	_srv        *server.Server
)

func setup() error {
	// add the following argument in go test command:
	// --httptest.serve=127.0.0.1:8080
	_httpServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_srv.ServeHTTP(w, r)
	}))

	store := memory.New()

	config := server.Config{
		IssuerURL:    "http://127.0.0.1:5556/dex",
		RedirectURL:  "http://127.0.0.1:8080/callback",
		ClientID:     "auth-service",
		ClientSecret: "ZXhhbXBsZS1hcHAtc2VjcmV0",
		Store:        store,
	}

	var err error
	_srv, err = server.NewServer(config)
	if err != nil {
		return err
	}
	return nil
}

func TestMain(m *testing.M) {
	err := setup()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to setup test server: %v", err)
		os.Exit(1)
	}
	defer _httpServer.Close()

	os.Exit(m.Run())
}

func TestUnauthorizedRequest(t *testing.T) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Get(_httpServer.URL)
	if err != nil {
		t.Fatal("failed to contact auth-service", err)
	}
	defer resp.Body.Close() // nolint
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("expected %v, got %v.", http.StatusFound, resp.StatusCode)
	}
}

/*
func TestAuthorizedRequests(t *testing.T) {
	url := httpServer.URL
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return errors.New("stooped after 10 redirects")
			}
			//req.SetBasicAuth("admin@example.com", "password")
			return nil
		},
	}
	resp, err := client.Get(url)
	if err != nil {
		t.Fatal("failed to contact auth-service", err)
	}
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("expected %v, got %v.", http.StatusFound, resp.StatusCode)
	}
}
*/

package server

import (
	"fmt"
	"github.com/fezho/oidc-auth-service/storage/memory"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestServer(t *testing.T, updateConfig func(c *Config)) (*httptest.Server, *Server) {
	var srv *Server
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		srv.ServeHTTP(w, r)
	}))

	store := memory.New()

	// TODO: call dex api to create client

	config := Config{
		IssuerURL:    "http://127.0.0.1:5556/dex",
		RPCEndpoint:  "127.0.0.1:5557",
		Address:      s.URL,
		ClientID:     "test-auth-app",
		ClientSecret: "test-auth-app-secret",
		Store:        store,
	}
	if updateConfig != nil {
		updateConfig(&config)
	}

	srv, err := NewServer(config)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("redirect url " + srv.oauth2Config.RedirectURL)
	_, err = srv.dexy.CreateClient(config.ClientID, config.ClientID, srv.oauth2Config.RedirectURL)
	if err != nil {
		t.Fatal(err)
	}

	return s, srv
}

func TestNewTestServer(t *testing.T) {
	httpServer, _ := newTestServer(t, nil)
	defer httpServer.Close()
}

// e2e test

func TestUnauthorizedRequest(t *testing.T) {
	httpServer, _ := newTestServer(t, nil)
	defer httpServer.Close()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	fmt.Printf("auth-service url: %s\n", httpServer.URL)
	resp, err := client.Get(httpServer.URL)
	if err != nil {
		t.Fatal("failed to contact auth-service", err)
	}
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("expected %v, got %v.", http.StatusFound, resp.StatusCode)
	}
}

func TestAuthCodeFlow(t *testing.T) {
	// TODO
}

func TestImplicitFlow(t *testing.T) {

}

// https://github.com/onelogin/openid-connect-dotnet-core-sample

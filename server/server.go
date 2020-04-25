package server

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/coreos/go-oidc"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

type Config struct {
	// Dex server address, Optional.
	DexAddress string
	// URL of the OpenID Connect issuer
	IssuerURL string
	// callback url for OpenID Connect Provider response.
	RedirectURL string
	// OAuth2 client ID of this application
	ClientID string
	// OAuth2 client secret of this application
	ClientSecret string
	// Scope specifies optional requested permissions
	Scopes []string
	// UsernameClaim is the JWT field to use as the user's username.
	UsernameClaim string
	// GroupsClaim, if specified, causes the OIDCAuthenticator to try to populate the user's
	// groups with an ID Token field. If the GroupsClaim field is present in an ID Token the value
	// must be a string or list of strings.
	GroupsClaim string
	// backend session store
	Store sessions.Store
	// CORS allowed origins
	AllowedOrigins []string
	// Whether to use AccessTypeOffline or not
	OfflineAccess bool
}

type Server struct {
	provider     *oidc.Provider
	store        sessions.Store
	oauth2Config *oauth2.Config

	usernameClaim string
	groupsClaim   string
	offlineAccess bool
	authCodeOpts  []oauth2.AuthCodeOption

	mux http.Handler
}

type UserIDOpts struct {
	Header      string
	TokenHeader string
	Prefix      string
	Claim       string
}

func NewServer(config Config) (*Server, error) {
	// Get callback, and root direct path from config.RedirectURL
	url, err := url.Parse(config.RedirectURL)
	if err != nil {
		return nil, fmt.Errorf("server: can't parse redirect URL %q", config.RedirectURL)
	}
	callback := path.Base(url.Path)
	dir := path.Dir(url.Path)

	client := http.DefaultClient
	if config.DexAddress != "" {
		client.Transport = NewDexRewriteURLRoundTripper(config.DexAddress, http.DefaultTransport)
	}
	ctx := oidc.ClientContext(context.Background(), client)
	provider, err := oidc.NewProvider(ctx, config.IssuerURL)
	if err != nil {
		return nil, errors.Errorf("server: can't get oidc provider %q: %v", config.IssuerURL, err)
	}

	// This is the only mandatory scope and will return a sub claim
	// which represents a unique identifier for the authenticated user.
	oidcScopes := append(config.Scopes, oidc.ScopeOpenID)

	s := &Server{
		provider: provider,
		oauth2Config: &oauth2.Config{
			RedirectURL:  config.RedirectURL,
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			Endpoint:     provider.Endpoint(),
			Scopes:       oidcScopes,
		},
		store:         config.Store,
		usernameClaim: config.UsernameClaim,
		groupsClaim:   config.GroupsClaim,
		offlineAccess: config.OfflineAccess,
	}

	router := mux.NewRouter()
	handleWithMethodGet := func(p string, f func(http.ResponseWriter,
		*http.Request)) {
		router.HandleFunc(path.Join(dir, p), f).Methods(http.MethodGet)
	}

	// Authorization redirect callback from OAuth2 auth flow.
	handleWithMethodGet(callback, s.callback)
	handleWithMethodGet("logout", s.logout)

	if config.OfflineAccess {
		s.authCodeOpts = append(s.authCodeOpts, oauth2.AccessTypeOffline)
		// TODO: review refresh_token api
		handleWithMethodGet("refresh_token", bearerTokenHandler(s.refreshToken))
	}

	// Handle health check
	router.Handle("/healthz", s.healthCheck(context.Background()))

	// Avoid root path being required twice from web browser
	router.HandleFunc("/favicon.ico", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	})

	// Auth check for root path
	router.PathPrefix("/").HandlerFunc(s.auth)

	s.mux = router
	if len(config.AllowedOrigins) > 0 {
		corsOption := handlers.AllowedOrigins(config.AllowedOrigins)
		s.mux = handlers.CORS(corsOption)(router)
	}

	return s, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

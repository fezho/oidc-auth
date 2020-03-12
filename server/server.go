package server

import (
	"context"
	"github.com/coreos/go-oidc"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
	"net/http"

	//log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type Config struct {
	// URL of the OpenID Connect issuer
	IssuerURL string
	// Callback URL for OAuth2 responses
	RedirectURI string // uri or url

	StaticDestinationURL string

	// OAuth2 client ID of this application
	ClientID string
	// OAuth2 client secret of this application
	ClientSecret string

	// Scope specifies optional requested permissions
	Scopes []string

	// whitelist uri list for skipping auth
	AuthWhitelistURI []string

	StorePath string

	SessionMaxAgeSeconds int

	UserIDOpts UserIDOpts

	// List of allowed origins for CORS requests on discovery, token and keys endpoint.
	// If none are indicated, CORS requests are disabled. Passing in "*" will allow any
	// domain.
	AllowedOrigins []string
}

type Server struct {
	provider             *oidc.Provider
	oauth2Config         *oauth2.Config
	store                sessions.Store
	staticDestination    string
	sessionMaxAgeSeconds int
	userIDOpts           UserIDOpts

	mux http.Handler
}

type UserIDOpts struct {
	Header      string
	TokenHeader string
	Prefix      string
	Claim       string
}

func NewServer(config *Config, store sessions.Store) (*Server, error) {
	// OIDC Discovery
	// TODO: retry with backoff
	provider, err := oidc.NewProvider(context.Background(), config.IssuerURL)
	if err != nil {
		return nil, errors.Errorf("failed to setup oidc provider %q: %v", config.IssuerURL, err)
	}

	oidcScopes := append(config.Scopes, oidc.ScopeOpenID)

	s := &Server{
		provider: provider,
		oauth2Config: &oauth2.Config{
			ClientID:     config.ClientSecret,
			ClientSecret: config.ClientSecret,
			Endpoint:     provider.Endpoint(),
			RedirectURL:  config.RedirectURI,
			Scopes:       oidcScopes,
		},
		store:                store,
		staticDestination:    config.StaticDestinationURL,
		sessionMaxAgeSeconds: config.SessionMaxAgeSeconds,
		userIDOpts:           config.UserIDOpts,
	}

	// register handlers
	router := mux.NewRouter()
	router.HandleFunc("/login", s.callback).Methods(http.MethodGet)
	router.HandleFunc("/logout", s.logout).Methods(http.MethodGet)
	router.PathPrefix("/").HandlerFunc(s.authenticate) // TODO: use authentication middleware https://github.com/gorilla/mux#middleware
	//router.PathPrefix("/api").Subrouter().Use()

	// set whitelist
	if len(config.AuthWhitelistURI) > 0 {
		router.Use(whitelistMiddleware(config.AuthWhitelistURI))
	}

	// set cors
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

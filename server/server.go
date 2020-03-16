package server

import (
	"context"
	"github.com/coreos/go-oidc"
	"github.com/fezho/oidc-auth-service/dex"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"net/http"
	//log "github.com/sirupsen/logrus"
)

type Config struct {
	// URL of the OpenID Connect issuer
	IssuerURL string
	// gRPC endpoint of dex server
	RPCEndpoint string
	// address of current web server
	Address string

	// OAuth2 client ID of this application
	ClientID string
	// OAuth2 client secret of this application
	ClientSecret string

	// Scope specifies optional requested permissions
	Scopes []string

	// uri whitelist for skipping auth
	URIWhitelist []string

	UserIDOpts UserIDOpts

	// backend session store
	Store sessions.Store

	// CORS allowed origins
	AllowedOrigins []string
}

type Server struct {
	provider             *oidc.Provider
	dexy                 *dex.Dexy
	store                sessions.Store
	oauth2Config         *oauth2.Config
	staticDestination    string
	sessionMaxAgeSeconds int
	userIDOpts           UserIDOpts
	redirectURL          string

	mux http.Handler
}

type UserIDOpts struct {
	Header      string
	TokenHeader string
	Prefix      string
	Claim       string
}

var callbackPath = "/login"

func NewServer(config Config) (*Server, error) {
	// OIDC Discovery
	// TODO: retry with backoff
	provider, err := oidc.NewProvider(context.Background(), config.IssuerURL)
	if err != nil {
		return nil, errors.Errorf("failed to setup oidc provider %q: %v", config.IssuerURL, err)
	}

	dexy, err := dex.New(config.RPCEndpoint)
	if err != nil {
		return nil, errors.Errorf("failed to create dex grpc client: %v", err)
	}

	oidcScopes := append(config.Scopes, oidc.ScopeOpenID)

	s := &Server{
		provider:    provider,
		dexy:        dexy,
		redirectURL: config.Address + callbackPath,
		oauth2Config: &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			Endpoint:     provider.Endpoint(),
			RedirectURL:  config.Address + callbackPath,
			Scopes:       oidcScopes,
		},
		store:      config.Store,
		userIDOpts: config.UserIDOpts,
	}

	// register handlers
	router := mux.NewRouter()
	// Authorization redirect callback from OAuth2 auth flow.
	router.HandleFunc(callbackPath, s.callback).Methods(http.MethodGet)
	router.HandleFunc("/logout", s.logout).Methods(http.MethodGet)

	router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		// do something
	})
	router.Use(s.authMiddleware())

	//authRouter := router.PathPrefix("/api").Subrouter()
	//authRouter.Use(s.authMiddleware())
	//authRouter.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	//	// Do something here
	//	fmt.Println("okay")
	//})

	// TODO: add session detail and refresh token api
	// https://github.com/Etiennera/go-ad-oidc/blob/741d9d275aa92d8e1243b68975dae36874b48026/activeDirectory_private.go#L101
	// https://developers.onelogin.com/openid-connect/api/refresh-session

	// set whitelist
	if len(config.URIWhitelist) > 0 {
		router.Use(whitelistMiddleware(config.URIWhitelist))
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

// TODO: use this func to replace oauth2Config
func (s *Server) genOauth2Config(r *http.Request) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     "",
		ClientSecret: "",
		Endpoint:     oauth2.Endpoint{},
		RedirectURL:  "",
		Scopes:       nil,
	}
}

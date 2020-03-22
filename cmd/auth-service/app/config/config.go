package config

import (
	"encoding/json"
	"fmt"
	"github.com/fezho/oidc-auth-service/storage"
	_ "github.com/fezho/oidc-auth-service/storage/bolt"
	_ "github.com/fezho/oidc-auth-service/storage/redis"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"os"
	"strings"
)

type Config struct {
	Web     Web     `json:"web"`
	OIDC    OIDC    `json:"oidc"`
	Storage Storage `json:"storage"`
	Logger  Logger  `json:"logger"`
}

func LoadConfigFromFile(file string) (*Config, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return LoadConfig(data)
}

func LoadConfig(data []byte) (*Config, error) {
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	return &c, nil
}

func (c Config) Validate() error {
	// Fast checks. Perform these first for a more responsive CLI.
	checks := []struct {
		bad    bool
		errMsg string
	}{
		{c.OIDC.Issuer == "", "no openID connect issuer specified"},
		{c.OIDC.RedirectURL == "", "no openID connect redirect url specified"},
		{c.OIDC.ClientID == "", "no openID connect client id specified"},
		{c.OIDC.ClientSecret == "", "no openID connect client secret specified"},
		{c.OIDC.UsernameClaim == "", "no openID connect user name claim specified"},
		{c.Storage.Config == nil, "no storage supplied in config file"},
		{c.Web.HTTP == "" && c.Web.HTTPS == "", "must supply a HTTP/HTTPS  address to listen on"},
		{c.Web.HTTPS != "" && c.Web.TLSCert == "", "no cert specified for HTTPS"},
		{c.Web.HTTPS != "" && c.Web.TLSKey == "", "no private key specified for HTTPS"},
	}

	var checkErrors []string
	for _, check := range checks {
		if check.bad {
			checkErrors = append(checkErrors, check.errMsg)
		}
	}
	if len(checkErrors) != 0 {
		return fmt.Errorf("invalid Config:\n\t-\t%s", strings.Join(checkErrors, "\n\t-\t"))
	}
	return nil
}

// Web is the config format for the HTTP server.
type Web struct {
	HTTP    string `json:"http"`
	HTTPS   string `json:"https"`
	TLSCert string `json:"tlsCert"`
	TLSKey  string `json:"tlsKey"`
	// List of allowed origins for CORS requests on discovery, token and keys endpoint.
	// If none are indicated, CORS requests are disabled. Passing in "*" will allow any domain.
	AllowedOrigins []string `json:"allowedOrigins"`
}

// OIDC is the config for authorization handlers with oidc provider
type OIDC struct {
	// URL of the OpenID Connect issuer
	// Required.
	Issuer string `json:"issuer"`
	// Redirect URL is the callback URL for OAuth2 responses, should be auth-service's URL+callback_path
	// Required.
	RedirectURL string `json:"redirectURL"`
	// OAuth2 client ID of this application
	// Required.
	ClientID string `json:"clientID"`
	// OAuth2 client secret of this application
	// Required.
	ClientSecret string `json:"clientSecret"`
	// Scope specifies optional requested permissions
	Scopes []string `json:"scopes"`
	// If your application needs to refresh access tokens when the user
	// is not present at the browser, then use offline.
	// Then the token response will include a refresh token.
	OfflineAccess bool `json:"offlineAccess"`
	// UsernameClaim is the JWT field to use as the user's username.
	// Required.
	UsernameClaim string `json:"usernameClaim"`
	// GroupsClaim, if specified, causes the OIDCAuthenticator to try to populate the user's
	// groups with an ID Token field. If the GroupsClaim field is present in an ID Token the value
	// must be a string or list of strings.
	GroupsClaim string `json:"groupsClaim"`
}

// Storage holds app's storage configuration.
type Storage struct {
	Type   storage.Type   `json:"type"`
	Config storage.Config `json:"config"`
}

// dynamically determine the type of the storage config.
func (s *Storage) UnmarshalJSON(b []byte) error {
	var store struct {
		Type   storage.Type    `json:"type"`
		Config json.RawMessage `json:"config"`
	}
	if err := json.Unmarshal(b, &store); err != nil {
		return fmt.Errorf("parse storage: %v", err)
	}

	cfg, err := storage.BuildStorageConfig(store.Type)
	if err != nil {
		return fmt.Errorf("unknown storage type %q", store.Type)
	}

	if len(store.Config) == 0 {
		return fmt.Errorf("no storage config found")
	}

	data := []byte(os.ExpandEnv(string(store.Config)))
	if err := json.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("parse storage config: %v", err)
	}

	*s = Storage{
		Type:   store.Type,
		Config: cfg,
	}
	return nil
}

// Logger holds configuration required to customize logging for dex.
type Logger struct {
	// Level sets logging level severity.
	Level string `json:"level"`
	// Format specifies the format to be used for logging.
	Format string `json:"format"`
}

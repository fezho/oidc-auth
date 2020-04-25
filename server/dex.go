package server

import (
	"net/http"
	"net/url"
	"strings"
)

// NewDexRewriteURLRoundTripper creates a new DexRewriteURLRoundTripper
func NewDexRewriteURLRoundTripper(dexAddr string, t http.RoundTripper) DexRewriteURLRoundTripper {
	if !strings.HasPrefix(dexAddr, "http://") && !strings.HasPrefix(dexAddr, "https://") {
		dexAddr = "http://" + dexAddr
	}

	dexURL, _ := url.Parse(dexAddr)
	return DexRewriteURLRoundTripper{
		DexURL: dexURL,
		T:      t,
	}
}

// DexRewriteURLRoundTripper is an HTTP RoundTripper to rewrite HTTP requests to the specified
// dex server address. This is used when requests Dex in same cluster to avoid from api gateway
// or external load balancer, which is not always permitted in firewalled/air-gapped networks.
type DexRewriteURLRoundTripper struct {
	DexURL *url.URL
	T      http.RoundTripper
}

func (s DexRewriteURLRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	// Exclude the .well-known path, in case to get error from oidc package:
	// oidc: issuer did not match the issuer returned by provider
	//if !strings.HasSuffix(r.URL.Path, "/.well-known/openid-configuration") {
	r.URL.Host = s.DexURL.Host
	r.URL.Scheme = s.DexURL.Scheme
	//}

	return s.T.RoundTrip(r)
}

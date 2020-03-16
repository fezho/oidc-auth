package server

import (
	"github.com/coreos/go-oidc"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"net/http"
	"strings"
)

const authSessionName = "oidc-auth-session"

// callback is the handler responsible for exchanging the auth_code and retrieving an id_token.
func (s *Server) callback(w http.ResponseWriter, r *http.Request) {
	logger := loggerForRequest(r)
	// TODO: add logger middleware

	// Get authorization code from authorization response.
	var authCode = r.FormValue("code")
	if len(authCode) == 0 {
		http.Error(w, "Missing url parameter: code", http.StatusBadRequest)
		return
	}

	var state = r.FormValue("state")
	if len(state) == 0 {
		http.Error(w, "Missing url parameter: state", http.StatusBadRequest)
		return
	}

	session, err := s.store.Get(r, authSessionName)
	if err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if nonce := session.Flashes("nonce"); len(nonce) == 0 || nonce[0].(string) != state {
		http.Error(w, "Forbidden", http.StatusForbidden)
		// session.Save(r, w) why?
		return
	}

	// Exchange the authorization code with {access, refresh, id}_token
	oauth2Token, err := s.oauth2Config.Exchange(r.Context(), authCode)
	if err != nil {
		http.Error(w, "failed to exchange authorization code with token", http.StatusInternalServerError)
		return
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		logger.Error("No id_token field available.")
		http.Error(w, "no id_token field in OAuth 2.0 token", http.StatusInternalServerError)
		return
	}

	// Verifying received ID token
	verifier := s.provider.Verifier(&oidc.Config{ClientID: s.oauth2Config.ClientID})
	_, err = verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		http.Error(w, "failed to verify ID token", http.StatusInternalServerError)
		return
	}

	// UserInfo endpoint to get claims
	claims := map[string]interface{}{}
	userInfo, err := s.provider.UserInfo(r.Context(), oauth2.StaticTokenSource(oauth2Token))
	if err != nil {
		http.Error(w, "failed to fetch user info", http.StatusInternalServerError)
		return
	}

	if err = userInfo.Claims(&claims); err != nil {
		logger.Println("Problem getting userinfo claims:", err.Error())
		http.Error(w, "failed to fetch user info claims", http.StatusInternalServerError)
		return
	}

	session.Values["userid"] = claims[s.userIDOpts.Claim].(string)
	session.Values["claims"] = claims
	session.Values["idtoken"] = rawIDToken
	session.Values["oauth2token"] = oauth2Token
	if err := session.Save(r, w); err != nil {
		logger.Errorf("Couldn't create user session: %v", err)
	}

	logger.Info("Login validated with ID token, redirecting.")

	f := session.Flashes("redirect_to")
	if len(f) > 0 {
		if err := session.Save(r, w); err != nil {
			logger.Errorf("failed to save user session: %v", err)
		}
		http.Redirect(w, r, f[0].(string), http.StatusFound)
	}
}

func (s *Server) refreshToken(w http.ResponseWriter, r *http.Request) {
	// 1. same with dex's example-app's MethodPost
	// 2. do in 	rawIDToken, ok := oauth2Token.Extra("id_token").(string) ... to session.Save
	// 3. should set some token in response header?
}

// logout is the handler responsible for revoking the user's session.
func (s *Server) logout(w http.ResponseWriter, r *http.Request) {

	logger := loggerForRequest(r)

	// Revoke user session.
	session, err := s.store.Get(r, authSessionName)
	if err != nil {
		logger.Errorf("Couldn't get user session: %v", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if session.IsNew {
		logger.Warn("Request doesn't have a valid session.")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	session.Options.MaxAge = -1
	if err := sessions.Save(r, w); err != nil {
		logger.Errorf("Couldn't delete user session: %v", err)
	}
	logger.Info("Successful logout.")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func loggerForRequest(r *http.Request) *log.Entry {
	return log.WithContext(r.Context()).WithFields(log.Fields{
		"ip":      getUserIP(r),
		"request": r.URL.String(),
	})
}

func getUserIP(r *http.Request) string {
	headerIP := r.Header.Get("X-Forwarded-For")
	if headerIP != "" {
		return headerIP
	}

	return strings.Split(r.RemoteAddr, ":")[0]
}

func whitelistMiddleware(whitelist []string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := loggerForRequest(r)
			// Check whitelist
			for _, prefix := range whitelist {
				if strings.HasPrefix(r.URL.Path, prefix) {
					logger.Infof("URI is whitelisted. Accepted without authorization.")
					w.WriteHeader(http.StatusOK)
					return
				}
			}

			// Pass down the request to the next middleware (or final handler)
			next.ServeHTTP(w, r)
		})
	}
}

func (s *Server) authMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := loggerForRequest(r)

			// Check header for auth information.
			// Adding it to a cookie to treat both cases uniformly.
			// This is also required by the gorilla/sessions package.
			// TODO(yanniszark): change to standard 'Authorization: Bearer <value>' header
			bearer := r.Header.Get("X-Auth-Token")
			if bearer != "" {
				r.AddCookie(&http.Cookie{
					Name:  authSessionName,
					Value: bearer,
				})
			}

			// Check if user session is valid
			session, err := s.store.Get(r, authSessionName)
			if err != nil {
				logger.Errorf("failed to get session: %v", err)
				http.Error(w, "InternalError", http.StatusInternalServerError)
				return
			}

			// User is logged in
			if !session.IsNew {
				// Add userid header
				// TODO: add to r.Header?
				userID := session.Values["userid"].(string)
				if userID != "" {
					w.Header().Set(s.userIDOpts.Header, s.userIDOpts.Prefix+userID)
				}
				if s.userIDOpts.TokenHeader != "" {
					w.Header().Set(s.userIDOpts.TokenHeader, session.Values["idtoken"].(string))
				}
				next.ServeHTTP(w, r)
			}

			// User is NOT logged in.
			// Initiate OIDC Flow with Authorization Request.
			nonce := uuid.New().String()
			session.Flashes("redirect_to", "nonce")
			session.AddFlash(r.URL.String(), "redirect_to")
			session.AddFlash(nonce, "nonce")
			if err = session.Save(r, w); err != nil {
				logger.Errorf("failed to save session: %v", err)
				http.Error(w, "InternalError", http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, s.oauth2Config.AuthCodeURL(nonce), http.StatusFound)
		})
	}
}

package server

import (
	"context"
	"github.com/coreos/go-oidc"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"net/http"
	"strings"
	"time"
)

const authSessionName = "oidc-auth-session"

// callback is the handler responsible for exchanging the auth_code and retrieving an id_token.
func (s *Server) callback(w http.ResponseWriter, r *http.Request) {
	log.Info("go into callback")

	// Get authorization code from authorization response.
	var authCode = r.FormValue("code")
	if len(authCode) == 0 {
		http.Error(w, "missing required parameter: code", http.StatusBadRequest)
		return
	}

	var state = r.FormValue("state")
	if len(state) == 0 {
		http.Error(w, "missing required parameter: state", http.StatusBadRequest)
		return
	}

	session, err := s.store.Get(r, authSessionName)
	if err != nil {
		log.Errorf("failed to get session: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if nonce := session.Flashes("nonce"); len(nonce) == 0 || nonce[0].(string) != state {
		http.Error(w, "access is unauthorized", http.StatusUnauthorized)
		return
	}
	redirect := session.Flashes("redirect_to")
	log.Infof("redirect: %v", redirect)

	// Exchange the authorization code with {access, refresh, id}_token
	oauth2Token, err := s.oauth2Config.Exchange(r.Context(), authCode)
	if err != nil {
		http.Error(w, "authentication failed", http.StatusUnauthorized)
		return
	}
	session.Values["refresh-token"] = oauth2Token.RefreshToken

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "authentication failed", http.StatusUnauthorized)
		return
	}

	if valid := s.authenticateToken(rawIDToken, session, w, r); !valid {
		return
	}

	log.Info("Login validated with ID token, redirecting.")

	if len(redirect) > 0 {
		http.Redirect(w, r, redirect[0].(string), http.StatusFound)
	}
}

// refreshToken refreshes the token in session
func (s *Server) refreshToken(w http.ResponseWriter, r *http.Request) {
	session, err := s.store.Get(r, authSessionName)
	if err != nil {
		log.Error(err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if session.IsNew {
		log.Error(err)
		http.Error(w, "authentication failed", http.StatusUnauthorized)
		return
	}
	refresh, ok := session.Values["refresh-token"].(string)
	if !ok {
		log.Error(err)
		http.Error(w, "authentication failed", http.StatusUnauthorized)
		return
	}
	t := &oauth2.Token{
		RefreshToken: refresh,
		Expiry:       time.Now().Add(-time.Hour),
	}
	oauth2Token, err := s.oauth2Config.TokenSource(context.Background(), t).Token()
	if err != nil {
		log.Error(err)
		http.Error(w, "authentication failed", http.StatusUnauthorized)
		return
	}
	session.Values["refresh-token"] = oauth2Token.RefreshToken
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		log.Error(err)
		http.Error(w, "authentication failed", http.StatusUnauthorized)
		return
	}

	s.authenticateToken(rawIDToken, session, w, r)
}

// logout is the handler responsible for revoking the user's session.
func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	// Revoke user session
	session, err := s.store.Get(r, authSessionName)
	if err != nil {
		log.Error(err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if session.IsNew {
		// It's a new session, redirect back to auth page
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	session.Options.MaxAge = -1
	if err := sessions.Save(r, w); err != nil {
		log.Errorf("Couldn't delete user session: %v", err)
	}

	// Redirect back to the auth page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) auth(w http.ResponseWriter, r *http.Request) {
	// Check if user session is valid
	log.Info("go into auth")
	session, err := s.store.Get(r, authSessionName)
	if err != nil {
		log.Error(err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	// User is logged in
	if !session.IsNew {
		if _, ok := session.Values["user_name"]; ok {
			s.login(session, w)
			return
		}
	}

	// User is NOT logged in
	// remove redirect_to & nonce first to make sure they are latest
	session.Flashes("redirect_to")
	session.Flashes("nonce")
	nonce := uuid.New().String()
	session.AddFlash(r.URL.String(), "redirect_to")
	log.Infof("redirect url: %s", r.URL.String())
	session.AddFlash(nonce, "nonce")
	if err = session.Save(r, w); err != nil {
		log.Errorf("failed to save session: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	authCodeURL := s.oauth2Config.AuthCodeURL(nonce, oauth2.ApprovalForce)

	// check the connector_id in request parameters
	connectorID := r.URL.Query().Get("connector_id")
	if connectorID != "" {
		authCodeURL = authCodeURL + "&connector_id=" + connectorID
	}

	http.Redirect(w, r, authCodeURL, http.StatusFound)
}

// authenticateToken verifies received ID token, extracts claims, save session
func (s *Server) authenticateToken(token string, session *sessions.Session, w http.ResponseWriter, r *http.Request) bool {
	verifier := s.provider.Verifier(&oidc.Config{ClientID: s.oauth2Config.ClientID})
	idToken, err := verifier.Verify(r.Context(), token)
	if err != nil {
		http.Error(w, "authentication failed", http.StatusUnauthorized)
		return false
	}

	var c claims
	if err := idToken.Claims(&c); err != nil {
		log.Errorf("failed to parse oidc claims: %v", err)
		http.Error(w, "authentication failed", http.StatusUnauthorized)
		return false
	}

	username, err := c.extractUsername(s.usernameClaim)
	if err != nil {
		log.Error(err)
		http.Error(w, "authentication failed", http.StatusUnauthorized)
	}
	session.Values["user_name"] = username
	if len(s.groupsClaim) > 0 {
		groups, err := c.extractGroups(s.groupsClaim)
		if err != nil {
			log.Error(err)
			http.Error(w, "authentication failed", http.StatusUnauthorized)
		}
		session.Values["user_groups"] = groups
	}

	session.Values["id_token"] = token
	if err := sessions.Save(r, w); err != nil {
		log.Error(err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return false
	}
	return true
}

// login sets user info into response header
func (s *Server) login(session *sessions.Session, w http.ResponseWriter) {
	// TODO: maybe this code should be optimized
	user, ok := session.Values["user_name"].(string)
	if ok {
		w.Header().Set("user_name", user)
	}
	if len(s.groupsClaim) > 0 {
		groups, ok := session.Values["user_groups"].(string)
		if ok {
			w.Header().Set("user_groups", groups)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func bearerTokenHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Call the next handler in the chain.
		auth := r.Header.Get("Authorization")
		if auth != "" {
			if !strings.HasPrefix(auth, "Bearer ") {
				http.Error(w, "bad bearer token", http.StatusBadRequest)
				return
			}
			bearerToken := strings.TrimPrefix(auth, "Bearer ")
			r.AddCookie(&http.Cookie{
				Name:  authSessionName,
				Value: bearerToken,
			})
		}

		next.ServeHTTP(w, r)
	}
}

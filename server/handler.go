package server

import (
	"github.com/coreos/go-oidc"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"strings"

	"net/http"
)

const userSessionCookie = "authservice_session"

func (s *Server) authenticate(w http.ResponseWriter, r *http.Request) {
	logger := loggerForRequest(r)

	// Check header for auth information.
	// Adding it to a cookie to treat both cases uniformly.
	// This is also required by the gorilla/sessions package.
	// TODO(yanniszark): change to standard 'Authorization: Bearer <value>' header
	bearer := r.Header.Get("X-Auth-Token")
	if bearer != "" {
		r.AddCookie(&http.Cookie{
			Name:   userSessionCookie,
			Value:  bearer,
			Path:   "/",
			MaxAge: 1,
		})
	}

	// Check if user session is valid
	session, err := s.store.Get(r, userSessionCookie)
	if err != nil {
		logger.Errorf("Couldn't get user session: %v", err)
		returnStatus(w, http.StatusInternalServerError, "Couldn't get user session.")
		return
	}
	// User is logged in
	if !session.IsNew {
		// Add userid header
		userID := session.Values["userid"].(string)
		if userID != "" {
			w.Header().Set(s.userIDOpts.Header, s.userIDOpts.Prefix+userID)
		}
		if s.userIDOpts.TokenHeader != "" {
			w.Header().Set(s.userIDOpts.TokenHeader, session.Values["idtoken"].(string))
		}
		returnStatus(w, http.StatusOK, "OK")
		return
	}

	// User is NOT logged in.
	// Initiate OIDC Flow with Authorization Request.
	// TODO: use session.AddFlash(url), id = session_id  https://github.com/Etiennera/go-ad-oidc/blob/master/auth.go
	state := newState(r.URL.String())
	id, err := state.save(s.store)
	if err != nil {
		logger.Errorf("Failed to save state in store: %v", err)
		returnStatus(w, http.StatusInternalServerError, "Failed to save state in store.")
		return
	}

	http.Redirect(w, r, s.oauth2Config.AuthCodeURL(id), http.StatusFound)
}

// callback is the handler responsible for exchanging the auth_code and retrieving an id_token.
func (s *Server) callback(w http.ResponseWriter, r *http.Request) {
	logger := loggerForRequest(r)

	// Get authorization code from authorization response.
	var authCode = r.FormValue("code")
	if len(authCode) == 0 {
		logger.Error("Missing url parameter: code")
		returnStatus(w, http.StatusBadRequest, "Missing url parameter: code")
		return
	}

	// Get state and:
	// 1. Confirm it exists in our memory.
	// 2. Get the original URL associated with it.
	var stateID = r.FormValue("state")
	if len(stateID) == 0 {
		logger.Error("Missing url parameter: state")
		returnStatus(w, http.StatusBadRequest, "Missing url parameter: state")
		return
	}

	// If state is loaded, then it's correct, as it is saved by its id.
	state, err := load(s.store, stateID)
	if err != nil {
		logger.Errorf("Failed to retrieve state from store: %v", err)
		returnStatus(w, http.StatusInternalServerError, "Failed to retrieve state.")
	}

	// Exchange the authorization code with {access, refresh, id}_token
	oauth2Token, err := s.oauth2Config.Exchange(r.Context(), authCode)
	if err != nil {
		logger.Errorf("Failed to exchange authorization code with token: %v", err)
		returnStatus(w, http.StatusInternalServerError, "Failed to exchange authorization code with token.")
		return
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		logger.Error("No id_token field available.")
		returnStatus(w, http.StatusInternalServerError, "No id_token field in OAuth 2.0 token.")
		return
	}

	// Verifying received ID token
	verifier := s.provider.Verifier(&oidc.Config{ClientID: s.oauth2Config.ClientID})
	_, err = verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		logger.Errorf("Not able to verify ID token: %v", err)
		returnStatus(w, http.StatusInternalServerError, "Unable to verify ID token.")
		return
	}

	// UserInfo endpoint to get claims
	claims := map[string]interface{}{}
	userInfo, err := s.provider.UserInfo(r.Context(), oauth2.StaticTokenSource(oauth2Token))
	if err != nil {
		logger.Errorf("Not able to fetch userinfo: %v", err)
		returnStatus(w, http.StatusInternalServerError, "Not able to fetch userinfo.")
		return
	}

	if err = userInfo.Claims(&claims); err != nil {
		logger.Println("Problem getting userinfo claims:", err.Error())
		returnStatus(w, http.StatusInternalServerError, "Not able to fetch userinfo claims.")
		return
	}

	// User is authenticated, create new session.
	session := sessions.NewSession(s.store, userSessionCookie)
	session.Options.MaxAge = s.sessionMaxAgeSeconds
	session.Options.Path = "/"

	session.Values["userid"] = claims[s.userIDOpts.Claim].(string)
	session.Values["claims"] = claims
	session.Values["idtoken"] = rawIDToken
	session.Values["oauth2token"] = oauth2Token
	if err := session.Save(r, w); err != nil {
		logger.Errorf("Couldn't create user session: %v", err)
	}

	logger.Info("Login validated with ID token, redirecting.")

	// Getting original destination from DB with state
	var destination = state.origURL
	if s.staticDestination != "" {
		destination = s.staticDestination
	}

	http.Redirect(w, r, destination, http.StatusFound)
}

// logout is the handler responsible for revoking the user's session.
func (s *Server) logout(w http.ResponseWriter, r *http.Request) {

	logger := loggerForRequest(r)

	// Revoke user session.
	session, err := s.store.Get(r, userSessionCookie)
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

func returnStatus(w http.ResponseWriter, statusCode int, errorMsg string) {
	w.WriteHeader(statusCode)
	w.Write([]byte(errorMsg))
}

func whitelistMiddleware(whitelist []string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := loggerForRequest(r)
			// Check whitelist
			for _, prefix := range whitelist {
				if strings.HasPrefix(r.URL.Path, prefix) {
					logger.Infof("URI is whitelisted. Accepted without authorization.")
					returnStatus(w, http.StatusOK, "OK")
					return
				}
			}

			// Pass down the request to the next middleware (or final handler)
			next.ServeHTTP(w, r)
		})
	}
}

package testutils

import (
	"github.com/fezho/oidc-auth/storage"
	"github.com/gorilla/securecookie"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func RunTestNew(t *testing.T, s *storage.Storage) {
	t.Run("New", func(t *testing.T) {
		// round 1
		req, _ := http.NewRequest("GET", "http://www.example.com", nil)
		session, err := s.New(req, "hello")
		if err != nil {
			t.Fatal("failed to create session, ", err)
		}
		if !session.IsNew {
			t.Fatalf("expected session is new, got false")
		}
		if session.ID != "" {
			t.Fatalf("expected session id is nil, got %s", session.ID)
		}
		// manually mark session is not new, since there's no Save operation, for test only
		session.IsNew = false

		// round 2 with same request, New returns new session
		//req, _ = http.NewRequest("GET", "http://www.example.com", nil)
		session, err = s.New(req, "hello")
		if err != nil {
			t.Fatal("failed to create session, ", err)
		}
		if !session.IsNew {
			t.Fatalf("expected session is new, got old")
		}
		if session.ID != "" {
			t.Fatalf("expected session id is nil, got %s", session.ID)
		}
		// manually mark session is not new, since there's no Save operation, for test only
		session.IsNew = false

		// round 3 with different request, New returns new session
		req, _ = http.NewRequest("GET", "http://www.example.com", nil)
		session, err = s.New(req, "hello")
		if err != nil {
			t.Fatal("failed to create session, ", err)
		}
		if !session.IsNew {
			t.Fatalf("expected session is new, got old")
		}
		if session.ID != "" {
			t.Fatalf("expected session id is nil, got %s", session.ID)
		}
	})
}

func RunTestGet(t *testing.T, s *storage.Storage) {
	t.Run("Get", func(t *testing.T) {
		// round 1
		req, _ := http.NewRequest("GET", "http://www.example.com", nil)
		session, err := s.Get(req, "hello")
		if err != nil {
			t.Fatal("failed to get session, ", err)
		}
		if !session.IsNew {
			t.Fatalf("expected session is new, got false")
		}
		if session.ID != "" {
			t.Fatalf("expected session id is nil, got %s", session.ID)
		}
		// manually mark session is not new, since there's no Save operation, for test only
		session.IsNew = false

		// round 2 with same request, get the same session of round 1 because of cache in request
		session, err = s.Get(req, "hello")
		if err != nil {
			t.Fatal("failed to get session, ", err)
		}
		if session.IsNew {
			t.Fatalf("expected session is old, got new")
		}
		if session.ID != "" {
			t.Fatalf("expected session id is nil, got %s", session.ID)
		}

		// round 3 with different request, get the different session of round 1
		req, _ = http.NewRequest("GET", "http://www.example.com", nil)
		session, err = s.Get(req, "hello")
		if err != nil {
			t.Fatal("failed to get session, ", err)
		}

		if !session.IsNew {
			t.Fatalf("expected session is new, got old")
		}
		if session.ID != "" {
			t.Fatalf("expected session id is nil, got %s", session.ID)
		}
	})
}

func RunTestSave(t *testing.T, s *storage.Storage) {
	t.Run("Save", func(t *testing.T) {
		// round 1 save session and check response's cookie
		req, _ := http.NewRequest("GET", "http://www.example.com", nil)
		session, err := s.New(req, "hello")
		if err != nil {
			t.Fatal("failed to create session", err)
		}

		session.Values["user"] = "tom"
		session.Values["tokens"] = map[string]interface{}{"token": "123456"}
		rsp := httptest.NewRecorder()
		err = session.Save(req, rsp)
		if err != nil {
			t.Fatal("failed to save session", err)
		}

		header := rsp.Header()
		cookies, ok := header["Set-Cookie"]
		if !ok || len(cookies) != 1 {
			t.Fatal("No cookies. Header:", header)
		}

		// round 2 check saved session by a request with same cookie of round 2
		req, _ = http.NewRequest("GET", "http://www.example.com", nil)
		req.Header.Add("Cookie", cookies[0])
		session, err = s.Get(req, "hello")
		if err != nil {
			t.Fatal("failed to get session, ", err)
		}
		if session.IsNew {
			t.Fatal("expected to get existed session, got new")
		}
		user := session.Values["user"]
		if user != "tom" {
			t.Fatalf("expected session value user:tom, got user:%s", user)
		}
		_, ok = session.Values["tokens"]
		if !ok {
			t.Fatal("expected session value map, got nothing")
		}

		// round 3 delete session, check response's cookie and new operation
		session.Options.MaxAge = 0
		rsp = httptest.NewRecorder()
		err = session.Save(req, rsp)
		if err != nil {
			t.Fatal("failed to delete session", err)
		}

		header = rsp.Header()
		cookies, ok = header["Set-Cookie"]
		if !ok || len(cookies) != 1 {
			t.Fatal("No cookies. Header:", header)
		}

		// round 4 use req with original cookie in round 2
		// use New operation to check session in order to not get it from cache
		session, err = s.New(req, "hello") // no cache
		if err != nil {
			t.Fatal("failed to get session, ", err)
		}
		if !session.IsNew {
			t.Fatalf("expected to get new session after being deleted")
		}
	})
}

func RunTestMaxAge(t *testing.T, s *storage.Storage) {
	t.Run("MaxAge", func(t *testing.T) {
		// round 1 save session and check response's cookie
		req, _ := http.NewRequest("GET", "http://www.example.com", nil)
		session, err := s.New(req, "hello")
		if err != nil {
			t.Fatal("failed to create session", err)
		}

		rsp := httptest.NewRecorder()
		err = session.Save(req, rsp)
		if err != nil {
			t.Fatal("failed to save session", err)
		}

		header := rsp.Header()
		cookies, ok := header["Set-Cookie"]
		if !ok || len(cookies) != 1 {
			t.Fatal("No cookies. Header:", header)
		}

		// set MaxAge
		s.MaxAge(1)
		// sleep 2 second to make sure session is expired
		time.Sleep(time.Second * 2)

		// round 2 use same req with Get method to get session
		// there's no cache in req, since Get is called first time
		req.Header.Add("Cookie", cookies[0])
		session, err = s.Get(req, "hello") // no cache
		if err == nil {
			t.Fatal("expected to get expired timestamp error, got nil")
		}
	})
}

func MockSessionConfig() storage.SessionConfig {
	key1 := string(securecookie.GenerateRandomKey(32))
	key2 := string(securecookie.GenerateRandomKey(32))
	return storage.SessionConfig{
		KeyPairs: []string{key1, key2},
		MaxAge:   1000,
	}
}

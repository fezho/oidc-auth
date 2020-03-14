package storage

import (
	"encoding/base32"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"net/http"
	"strings"
)

// Storage is a custom session store which provides an abstraction of
// common session store operations for multiple Key/Value databases.
type Storage struct {
	// conn is the underlying db to do load/save/delete operations
	conn Conn
	// codecs encode&decode session to/from cookie, it also checks MaxAge in `DecodeMulti` method
	codecs []securecookie.Codec
	// options stores configuration for a session
	options *sessions.Options
}

// Conn is the interface for underlying persistent database
type Conn interface {
	// Load reads the session from the database.
	// returns true if there is a session data in DB
	Load(session *sessions.Session) (bool, error)
	// Save stores the session in the database.
	Save(session *sessions.Session) error
	// Delete removes keys from the database if MaxAge<0
	Delete(session *sessions.Session) error
	// Close closes the database.
	Close() error
}

func New(conn Conn, config *SessionConfig) *Storage {
	s := &Storage{
		codecs: securecookie.CodecsFromPairs(config.KeyPairs...),
		options: &sessions.Options{
			Path:   "/",
			MaxAge: config.MaxAge,
		},
		conn: conn,
	}

	s.MaxAge(s.options.MaxAge)
	return s
}

// Get returns a session for the given name after adding it to the registry.
//
// It returns a new session if the sessions doesn't exist. Access IsNew on
// the session to check if it is an existing session or a new one.
//
// It returns a new session and an error if the session exists but could
// not be decoded or be expired.
func (s *Storage) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

// New returns a session for the given name without adding it to the registry.
//
// The difference between New() and Get() is that calling New() twice will
// decode the session data twice, while Get() registers and reuses the same
// decoded session after the first call. Get() calls New() internally if
// there's no data in cache.
func (s *Storage) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(s, name)
	opts := *s.options
	session.Options = &opts
	session.IsNew = true
	var err error
	if c, errCookie := r.Cookie(name); errCookie == nil {
		err = securecookie.DecodeMulti(name, c.Value, &session.ID, s.codecs...)
		if err == nil {
			ok, err := s.conn.Load(session)
			session.IsNew = !(err == nil && ok) // not new if no error and data available
		}
	}
	return session, err
}

// Save saves a single session to the underlying database
// and save the encoded session id to cookie of the response.
//
// If the options.MaxAge of the session is <= 0 then the session will be
// deleted from the store path. With this process it enforces the properly
// session cookie handling so no need to trust in the cookie management in the
// web browser.
func (s *Storage) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	// Delete if max-age is <= 0
	if session.Options.MaxAge <= 0 {
		if err := s.conn.Delete(session); err != nil {
			return err
		}
		http.SetCookie(w, sessions.NewCookie(session.Name(), "", session.Options))
		return nil
	}

	// encode id to use alphanumeric characters only for internal db usage.
	if session.ID == "" {
		session.ID = strings.TrimRight(
			base32.StdEncoding.EncodeToString(
				securecookie.GenerateRandomKey(32)), "=")
	}
	if err := s.conn.Save(session); err != nil {
		return err
	}
	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, s.codecs...)
	if err != nil {
		return err
	}
	http.SetCookie(w, sessions.NewCookie(session.Name(), encoded, session.Options))
	return nil
}

// MaxAge sets the maximum age for the store and the underlying cookie
// implementation. Individual sessions can be deleted by setting options.MaxAge
// = -1 for that session.
func (s *Storage) MaxAge(age int) {
	s.options.MaxAge = age

	// Set the maxAge for each securecookie instance.
	for _, codec := range s.codecs {
		if sc, ok := codec.(*securecookie.SecureCookie); ok {
			sc.MaxAge(age)
		}
	}
}

func (s *Storage) Close() error {
	return s.conn.Close()
}

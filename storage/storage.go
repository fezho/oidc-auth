package storage

import (
	"encoding/base32"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"net/http"
	"strings"
)

// Storage is a base storage for custom session stores.
type Storage struct {
	conn    Conn
	Codecs  []securecookie.Codec
	Options *sessions.Options
}

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

// Config is the basic storage options
type Config struct {
	Serializer securecookie.Serializer
	// Session Max-Age attribute present and given in seconds.
	MaxAge int
	// KeyPairs are used to generate securecookie.Codec,
	// Should not change them after application is started,
	// otherwise previously issued cookies will not be able to be decoded.
	// Can be created using securecookie.GenerateRandomKey()
	KeyPairs [][]byte
}

func New(conn Conn, config *Config) *Storage {
	s := &Storage{
		Codecs: securecookie.CodecsFromPairs(config.KeyPairs...),
		Options: &sessions.Options{
			Path:   "/",
			MaxAge: config.MaxAge,
		},
		conn: conn,
	}

	s.MaxAge(s.Options.MaxAge)
	return s
}

// Get returns a session for the given name after adding it to the registry.
//
// It returns a new session if the sessions doesn't exist. Access IsNew on
// the session to check if it is an existing session or a new one.
//
// It returns a new session and an error if the session exists but could
// not be decoded.
func (s *Storage) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

// New returns a session for the given name without adding it to the registry.
//
// The difference between New() and Get() is that calling New() twice will
// decode the session data twice, while Get() registers and reuses the same
// decoded session after the first call.
func (s *Storage) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(s, name)
	opts := *s.Options
	session.Options = &opts
	session.IsNew = true
	var err error
	if c, errCookie := r.Cookie(name); errCookie == nil {
		err = securecookie.DecodeMulti(name, c.Value, &session.ID, s.Codecs...)
		if err == nil {
			ok, err := s.conn.Load(session)
			// TODO: !ok is enough
			session.IsNew = !(err == nil && ok) // not new if no error and data available
		}
	}
	return session, err
}

// Save adds a single session to the response.
//
// If the Options.MaxAge of the session is <= 0 then the session will be
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
	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, s.Codecs...)
	if err != nil {
		return err
	}
	http.SetCookie(w, sessions.NewCookie(session.Name(), encoded, session.Options))
	return nil
}

// MaxAge sets the maximum age for the store and the underlying cookie
// implementation. Individual sessions can be deleted by setting Options.MaxAge
// = -1 for that session.
func (s *Storage) MaxAge(age int) {
	s.Options.MaxAge = age

	// Set the maxAge for each securecookie instance.
	for _, codec := range s.Codecs {
		if sc, ok := codec.(*securecookie.SecureCookie); ok {
			sc.MaxAge(age)
		}
	}
}

func (s *Storage) Close() error {
	return s.conn.Close()
}

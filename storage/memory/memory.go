package memory

import (
	"sync"
	"time"

	"github.com/gorilla/sessions"

	"github.com/fezho/oidc-auth/storage"
	"github.com/fezho/oidc-auth/storage/internal"
)

type Config struct {
	storage.SessionConfig `json:",inline"`
}

func init() {
	storage.AddConfigBuilder(storage.MEMORY, func() storage.Config { return new(Config) })
}

func (c *Config) Open() (*storage.Storage, error) {
	conn := &memoryConn{
		sessions: make(map[string]valueType),
	}
	return storage.New(conn, c.SessionConfig), nil
}

// New initiate a memory storage, for test only
func New() *storage.Storage {
	cfg := Config{}
	s, _ := cfg.Open()
	return s
}

type memoryConn struct {
	mu sync.RWMutex

	sessions map[string]valueType
}

type valueType struct {
	data []byte
	ttl  int64
}

func (m *memoryConn) Load(session *sessions.Session) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	value, ok := m.sessions[session.ID]
	if !ok || isExpired(value.ttl) {
		return false, nil
	}

	return true, internal.Decode(value.data, session)
}

func (m *memoryConn) Save(session *sessions.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := internal.Encode(session)
	if err != nil {
		return err
	}

	value := valueType{
		data: data,
		ttl:  time.Now().UTC().Unix() + int64(session.Options.MaxAge),
	}
	m.sessions[session.ID] = value

	return nil
}

func (m *memoryConn) Delete(session *sessions.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.sessions, session.ID)
	return nil
}

func (m *memoryConn) Close() error {
	return nil
}

func isExpired(ttl int64) bool {
	return ttl > 0 && ttl <= time.Now().UTC().Unix()
}

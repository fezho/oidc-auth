package memory

import (
	"github.com/fezho/oidc-auth-service/storage"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"sync"
	"time"
)

type Config struct {
	// The in-memory implementation has no config.
	*storage.SessionConfig `json:",inline"`
}

func init() {
	storage.AddConfigBuilder(storage.MEMORY, func() storage.Config { return new(Config) })
}

func (c *Config) Open() (*storage.Storage, error) {
	err := c.SessionConfig.Unmarshal()
	if err != nil {
		return nil, err
	}

	conn := &memoryConn{
		sessions:   make(map[string]valueType),
		serializer: c.Serializer,
	}
	return storage.New(conn, c.SessionConfig), nil
}

type memoryConn struct {
	mu sync.RWMutex

	sessions   map[string]valueType
	serializer securecookie.Serializer
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

	return true, m.serializer.Deserialize(value.data, &session.Values)
}

func (m memoryConn) Save(session *sessions.Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := m.serializer.Serialize(session.Values)
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

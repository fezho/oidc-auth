package memory

import (
	"github.com/fezho/oidc-auth-service/storage"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"sync"
	"time"
)

type Config struct {
	// The in memory implementation has no config.
}

func (r *Config) Open(config *storage.Config) (*storage.Storage, error) {
	conn := &memoryConn{
		sessions:   make(map[string]valueType),
		serializer: config.Serializer,
	}
	return storage.New(conn, config), nil
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
		ttl:  time.Now().Unix() + int64(session.Options.MaxAge),
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
	return ttl > 0 && ttl <= time.Now().Unix()
}

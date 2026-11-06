package ratchet

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"go.mau.fi/libsignal/keys/identity"
	"go.mau.fi/libsignal/logger"
	"go.mau.fi/libsignal/protocol"
	"go.mau.fi/libsignal/serialize"
	"go.mau.fi/libsignal/session"
	"go.mau.fi/libsignal/state/record"
	"go.mau.fi/libsignal/util/keyhelper"
)

type Engine struct {
	mu         sync.Mutex
	store      *memoryStore
	bundles    map[string][]byte
	nextPreKey uint32
}

var configureLogger sync.Once

type silentLogger struct{}

func (silentLogger) Debug(string, string)   {}
func (silentLogger) Info(string, string)    {}
func (silentLogger) Warning(string, string) {}
func (silentLogger) Error(string, string)   {}
func (silentLogger) Configure(string)       {}

func New() (*Engine, error) {
	configureLogger.Do(func() {
		var sink logger.Loggable = silentLogger{}
		logger.Setup(&sink)
	})
	ik, err := keyhelper.GenerateIdentityKeyPair()
	if err != nil {
		return nil, fmt.Errorf("generate identity: %w", err)
	}
	serializer := serialize.NewProtoBufSerializer()
	spk, err := keyhelper.GenerateSignedPreKey(ik, 1, serializer.SignedPreKeyRecord)
	if err != nil {
		return nil, fmt.Errorf("generate signed prekey: %w", err)
	}
	keys, err := keyhelper.GeneratePreKeys(1, 100, serializer.PreKeyRecord)
	if err != nil {
		return nil, fmt.Errorf("generate prekeys: %w", err)
	}
	s := &memoryStore{
		ik: ik, registration: keyhelper.GenerateRegistrationID(), serializer: serializer,
		identities: make(map[string]*identity.Key), sessions: make(map[string][]byte),
		prekeys: make(map[uint32]*record.PreKey), signed: map[uint32]*record.SignedPreKey{1: spk},
	}
	for _, key := range keys {
		s.prekeys[key.ID().Value] = key
	}
	return &Engine{store: s, bundles: make(map[string][]byte), nextPreKey: 1}, nil
}

func (e *Engine) HasSession(remote string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	_, ok := e.store.sessions[protocol.NewSignalAddress(remote, 1).String()]
	return ok
}

func (e *Engine) builder(remote string) (*session.Builder, *protocol.SignalAddress) {
	address := protocol.NewSignalAddress(remote, 1)
	return session.NewBuilder(e.store, e.store, e.store, e.store, address, e.store.serializer), address
}

type memoryStore struct {
	ik           *identity.KeyPair
	registration uint32
	serializer   *serialize.Serializer
	identities   map[string]*identity.Key
	sessions     map[string][]byte
	prekeys      map[uint32]*record.PreKey
	signed       map[uint32]*record.SignedPreKey
}

func (s *memoryStore) GetIdentityKeyPair() *identity.KeyPair { return s.ik }
func (s *memoryStore) GetLocalRegistrationID() uint32        { return s.registration }
func (s *memoryStore) SaveIdentity(_ context.Context, address *protocol.SignalAddress, key *identity.Key) error {
	s.identities[address.String()] = key
	return nil
}
func (s *memoryStore) IsTrustedIdentity(_ context.Context, address *protocol.SignalAddress, key *identity.Key) (bool, error) {
	saved := s.identities[address.String()]
	return saved == nil || bytes.Equal(saved.Serialize(), key.Serialize()), nil
}
func (s *memoryStore) LoadSession(_ context.Context, address *protocol.SignalAddress) (*record.Session, error) {
	if data, ok := s.sessions[address.String()]; ok {
		return record.NewSessionFromBytes(data, s.serializer.Session, s.serializer.State)
	}
	return record.NewSession(s.serializer.Session, s.serializer.State), nil
}
func (s *memoryStore) StoreSession(_ context.Context, address *protocol.SignalAddress, r *record.Session) error {
	s.sessions[address.String()] = r.Serialize()
	return nil
}
func (s *memoryStore) ContainsSession(_ context.Context, address *protocol.SignalAddress) (bool, error) {
	_, ok := s.sessions[address.String()]
	return ok, nil
}
func (s *memoryStore) GetSubDeviceSessions(_ context.Context, _ string) ([]uint32, error) {
	return nil, nil
}
func (s *memoryStore) DeleteSession(_ context.Context, address *protocol.SignalAddress) error {
	delete(s.sessions, address.String())
	return nil
}
func (s *memoryStore) DeleteAllSessions(_ context.Context) error { clear(s.sessions); return nil }
func (s *memoryStore) LoadPreKey(_ context.Context, id uint32) (*record.PreKey, error) {
	key := s.prekeys[id]
	if key == nil {
		return nil, fmt.Errorf("prekey %d is unavailable", id)
	}
	return key, nil
}
func (s *memoryStore) StorePreKey(_ context.Context, id uint32, key *record.PreKey) error {
	s.prekeys[id] = key
	return nil
}
func (s *memoryStore) ContainsPreKey(_ context.Context, id uint32) (bool, error) {
	_, ok := s.prekeys[id]
	return ok, nil
}
func (s *memoryStore) RemovePreKey(_ context.Context, id uint32) error {
	delete(s.prekeys, id)
	return nil
}
func (s *memoryStore) LoadSignedPreKey(_ context.Context, id uint32) (*record.SignedPreKey, error) {
	key := s.signed[id]
	if key == nil {
		return nil, fmt.Errorf("signed prekey %d is unavailable", id)
	}
	return key, nil
}
func (s *memoryStore) LoadSignedPreKeys(_ context.Context) ([]*record.SignedPreKey, error) {
	keys := make([]*record.SignedPreKey, 0, len(s.signed))
	for _, key := range s.signed {
		keys = append(keys, key)
	}
	return keys, nil
}
func (s *memoryStore) StoreSignedPreKey(_ context.Context, id uint32, key *record.SignedPreKey) error {
	s.signed[id] = key
	return nil
}
func (s *memoryStore) ContainsSignedPreKey(_ context.Context, id uint32) (bool, error) {
	_, ok := s.signed[id]
	return ok, nil
}
func (s *memoryStore) RemoveSignedPreKey(_ context.Context, id uint32) error {
	delete(s.signed, id)
	return nil
}

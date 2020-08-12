package session

import (
	"context"
	"sync"
)

// mockSessionManager implements the needed functionality for the users resolver.
type MockSessionManager struct {
	keys map[string]interface{}
	lock *sync.Mutex
}

// NewMockSessionManager returns a pointer to a mockSessionManager instance.
func NewMockSessionManager() *MockSessionManager {
	return &MockSessionManager{
		keys: make(map[string]interface{}),
		lock: &sync.Mutex{},
	}
}

// Put puts a value into the session.
func (s *MockSessionManager) Put(ctx context.Context, key string, val interface{}) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.keys[key] = val
}

// RenewToken renews the user's session token.
func (s *MockSessionManager) RenewToken(ctx context.Context) error {
	// not implemented
	return nil
}

package manager

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"

	"github.com/rs/zerolog/log"
)

type AuthStore struct {
	mu          sync.RWMutex
	credentials map[string]string
}

// NewAuthStore creates a new instance of AuthStore.
func NewAuthStore() *AuthStore {
	log.Debug().Msg("Creating new AuthStore")
	return &AuthStore{
		credentials: map[string]string{},
	}
}

// hashPassword hashes a given password using SHA-256.
func hashPassword(password string) string {
	hash := sha256.New()
	hash.Write([]byte(password))
	return hex.EncodeToString(hash.Sum(nil))
}

// Add adds a new username-password combination to the store.
func (store *AuthStore) Add(username, password string) {
	log.Info().Msgf("Adding credentials for user: %s", username)
	store.mu.Lock()
	defer store.mu.Unlock()
	store.credentials[username] = hashPassword(password)
}

// Remove removes a username-password combination from the store.
func (store *AuthStore) Remove(username string) {
	log.Info().Msgf("Removing credentials for user: %s", username)
	store.mu.Lock()
	defer store.mu.Unlock()
	delete(store.credentials, username)
}

// Validate checks if a username-password combination is present in the store.
func (store *AuthStore) Validate(username, password string) bool {
	log.Info().Msgf("Validating credentials for user: %s", username)
	store.mu.RLock()
	defer store.mu.RUnlock()
	hashedPassword, exists := store.credentials[username]
	if !exists {
		return false
	}
	return hashedPassword == hashPassword(password)
}

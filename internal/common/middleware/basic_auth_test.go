package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mockAuthenticator struct {
	validCredentials map[string]string
}

func (m *mockAuthenticator) Validate(username, password string) bool {
	pass, ok := m.validCredentials[username]
	return ok && pass == password
}

func TestBasicAuth_Success(t *testing.T) {
	// Mock Authenticator with valid credentials
	validator := &mockAuthenticator{
		validCredentials: map[string]string{"user": "pass"},
	}

	// Create a test handler that will only be called on successful authentication
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Create the BasicAuth middleware with the mock Authenticator
	basicAuthMiddleware := BasicAuth("test-realm", validator)

	// Wrap the final handler with the middleware
	handler := basicAuthMiddleware(finalHandler)

	// Create a test request with valid basic auth headers
	req := httptest.NewRequest("GET", "/", nil)
	req.SetBasicAuth("user", "pass")
	w := httptest.NewRecorder()

	// Serve the request
	handler.ServeHTTP(w, req)

	// Check if the response status is OK and the body contains "success"
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if strings.TrimSpace(w.Body.String()) != "success" {
		t.Errorf("expected body %q, got %q", "success", w.Body.String())
	}
}

func TestBasicAuth_Failure_InvalidCredentials(t *testing.T) {
	// Mock Authenticator with valid credentials
	validator := &mockAuthenticator{
		validCredentials: map[string]string{"user": "pass"},
	}

	// Create the BasicAuth middleware with the mock Authenticator
	basicAuthMiddleware := BasicAuth("test-realm", validator)

	// Create a test handler that will not be called on failed authentication
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Wrap the final handler with the middleware
	handler := basicAuthMiddleware(finalHandler)

	// Create a test request with invalid basic auth headers
	req := httptest.NewRequest("GET", "/", nil)
	req.SetBasicAuth("user", "wrongpass")
	w := httptest.NewRecorder()

	// Serve the request
	handler.ServeHTTP(w, req)

	// Check if the response status is Unauthorized
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	// Check if the WWW-Authenticate header is set correctly
	authHeader := w.Header().Get("WWW-Authenticate")
	expectedAuthHeader := `Basic realm="test-realm"`
	if authHeader != expectedAuthHeader {
		t.Errorf("expected WWW-Authenticate header %q, got %q", expectedAuthHeader, authHeader)
	}
}

func TestBasicAuth_Failure_NoCredentials(t *testing.T) {
	// Mock Authenticator with valid credentials
	validator := &mockAuthenticator{
		validCredentials: map[string]string{"user": "pass"},
	}

	// Create the BasicAuth middleware with the mock Authenticator
	basicAuthMiddleware := BasicAuth("test-realm", validator)

	// Create a test handler that will not be called on failed authentication
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Wrap the final handler with the middleware
	handler := basicAuthMiddleware(finalHandler)

	// Create a test request with no basic auth headers
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	// Serve the request
	handler.ServeHTTP(w, req)

	// Check if the response status is Unauthorized
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	// Check if the WWW-Authenticate header is set correctly
	authHeader := w.Header().Get("WWW-Authenticate")
	expectedAuthHeader := `Basic realm="test-realm"`
	if authHeader != expectedAuthHeader {
		t.Errorf("expected WWW-Authenticate header %q, got %q", expectedAuthHeader, authHeader)
	}
}

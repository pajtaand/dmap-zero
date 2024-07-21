package manager

import (
	"fmt"
	"testing"
)

func TestAuthStore_AddAndValidate(t *testing.T) {
	authStore := NewAuthStore()

	authStore.Add("user1", "password1")
	authStore.Add("user2", "password2")

	tests := []struct {
		username string
		password string
		expected bool
	}{
		{"user1", "password1", true},
		{"user2", "password2", true},
		{"user1", "wrongpassword", false},
		{"user3", "password3", false},
	}

	for ind, test := range tests {
		t.Run(fmt.Sprintf("Test_%d", ind), func(t *testing.T) {
			result := authStore.Validate(test.username, test.password)
			if result != test.expected {
				t.Errorf("Validate(%q, %q) = %v; expected %v", test.username, test.password, result, test.expected)
			}
		})
	}
}

func TestAuthStore_Remove(t *testing.T) {
	authStore := NewAuthStore()

	authStore.Add("user1", "password1")
	authStore.Add("user2", "password2")

	authStore.Remove("user1")

	if authStore.Validate("user1", "password1") {
		t.Errorf("Validate(%q, %q) = true; expected false after removal", "user1", "password1")
	}

	if !authStore.Validate("user2", "password2") {
		t.Errorf("Validate(%q, %q) = false; expected true for remaining user", "user2", "password2")
	}
}

func TestAuthStore_RemoveNonexistentUser(t *testing.T) {
	authStore := NewAuthStore()

	authStore.Add("user1", "password1")

	authStore.Remove("user2") // Removing a non-existent user

	if !authStore.Validate("user1", "password1") {
		t.Errorf("Validate(%q, %q) = false; expected true after removing non-existent user", "", "password1")
	}
}

func TestAuthStore_HashPassword(t *testing.T) {
	tests := []struct {
		password     string
		expectedHash string
	}{
		{
			password:     "password123",
			expectedHash: "ef92b778bafe771e89245b89ecbc08a44a4e166c06659911881f383d4473e94f",
		},
		{
			password:     "secret",
			expectedHash: "2bb80d537b1da3e38bd30361aa855686bde0eacd7162fef6a25fe97bf527a25b",
		},
		{
			password:     "golang",
			expectedHash: "d754ed9f64ac293b10268157f283ee23256fb32a4f8dedb25c8446ca5bcb0bb3",
		},
		{
			password:     "password",
			expectedHash: "5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8",
		},
	}

	for ind, test := range tests {
		t.Run(fmt.Sprintf("Test_%d", ind), func(t *testing.T) {
			hashed := hashPassword(test.password)
			if hashed != test.expectedHash {
				t.Errorf("hashPassword(%q) = %v; expected %v", test.password, hashed, test.expectedHash)
			}
		})
	}
}

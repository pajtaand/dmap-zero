package utils

import "context"

type contextKey string

const (
	UserKey contextKey = "user"
)

// SetUser adds a user to the context.
func SetUser(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserKey, userID)
}

// GetUser retrieves the user from the context.
func GetUser(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserKey).(string)
	return userID, ok
}

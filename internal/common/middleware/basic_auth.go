// this middleware is based on https://github.com/go-chi/chi/blob/master/middleware/basic_auth.go
package middleware

import (
	"fmt"
	"net/http"

	"github.com/pajtaand/dmap-zero/internal/common/constants"
	"github.com/pajtaand/dmap-zero/internal/common/utils"
	"github.com/rs/zerolog"
)

type Authenticator interface {
	Validate(username, password string) bool
}

// BasicAuth implements a simple middleware handler for adding basic http auth to a route.
func BasicAuth(realm string, validator Authenticator) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()
			if !ok {
				basicAuthFailed(w, realm)
				return
			}

			if !validator.Validate(user, pass) {
				basicAuthFailed(w, realm)
				return
			}

			ctx := r.Context()

			// add user to logger
			l := zerolog.Ctx(ctx)
			l.UpdateContext(func(c zerolog.Context) zerolog.Context {
				return c.Str(constants.LoggerKeyRequestUser, user)
			})

			// add user to context
			ctx = utils.SetUser(ctx, user)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

func basicAuthFailed(w http.ResponseWriter, realm string) {
	w.Header().Add("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, realm))
	w.WriteHeader(http.StatusUnauthorized)
}

package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/vardius/gorouter/v4"

	"github.com/vardius/go-api-boilerplate/pkg/application"
	"github.com/vardius/go-api-boilerplate/pkg/errors"
	"github.com/vardius/go-api-boilerplate/pkg/http/response"
	"github.com/vardius/go-api-boilerplate/pkg/log"
)

// Recover middleware recovers from panic
func Recover(logger *log.Logger) gorouter.MiddlewareFunc {
	m := func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Critical(r.Context(), "[HTTP] Recovered in %v\n%s\n", rec, debug.Stack())

					appErr := errors.Wrap(fmt.Errorf("%w: recovered from panic", application.ErrInternal))

					if err := response.JSONError(r.Context(), w, appErr); err != nil {
						logger.Critical(r.Context(), "[HTTP] Errors while sending response after panic %v\n", err)
					}
				}
			}()

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}

	return m
}

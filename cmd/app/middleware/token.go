package middleware

import (
	"context"
	"errors"
	"net/http"
)

var ErrNoAuthentication = errors.New("no authentication")
var authenticationContextKey = &contextKey{"authentication context"}
var folderKey = &contextKey{"folder"}
var activeKey = &contextKey{"active"}

type contextKey struct {
	name string
}

func (c *contextKey) String() string {
	return c.name
}

type IDFunc func(ctx context.Context, token string) (int64, string, error, bool)

func Authenticate(idFunc IDFunc) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

			token := ""
			t, err := request.Cookie("session")
			if err == nil {
				token = t.Value
			}
			var id int64
			var folder string
			var active bool
			// Отклаючает мидлваре,только в целях отладки и тестирования. Что я делаю?
			if request.URL.String() != "/getcode" {

				id, folder, err, active = idFunc(request.Context(), token)
				if err != nil {
					http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}

			}
			ctx := context.WithValue(request.Context(), authenticationContextKey, id)
			request = request.WithContext(ctx)
			ctx = context.WithValue(request.Context(), folderKey, folder)
			request = request.WithContext(ctx)
			ctx = context.WithValue(request.Context(), activeKey, active)
			request = request.WithContext(ctx)

			handler.ServeHTTP(writer, request)
		})
	}
}

func Authentication(ctx context.Context) (int64, error) {
	if value, ok := ctx.Value(authenticationContextKey).(int64); ok {
		return value, nil
	}
	return 0, ErrNoAuthentication
}

func Folder(ctx context.Context) (string, error) {
	if value, ok := ctx.Value(folderKey).(string); ok {
		return value, nil
	}
	return "", ErrNoAuthentication
}

func UserActive(ctx context.Context) (bool, error) {
	if value, ok := ctx.Value(activeKey).(bool); ok {
		return value, nil
	}
	return false, ErrNoAuthentication
}

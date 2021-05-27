package http

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func MakeServer(logger *zap.Logger) (http.Handler, error) {
	r := mux.NewRouter()
	r.Use(recoveryMiddleware(logger))
	r.Handle("/ping", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		//nolint:errcheck // yeah lets just ignore that
		rw.Write([]byte("pong"))
	}))

	return r, nil
}

func recoveryMiddleware(logger *zap.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			defer func() {
				if r := recover(); r != nil {
					logger.Panic(fmt.Sprintf("http recovery middleware caught a panic: %+v", r))
				}
			}()

			next.ServeHTTP(rw, r)
		})
	}
}

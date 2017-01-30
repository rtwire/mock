package service

import "net/http"

type middleware []func(http.Handler) http.Handler

func (mw middleware) Handler(handler http.HandlerFunc) http.Handler {
	h := http.Handler(handler)
	for i := len(mw) - 1; i >= 0; i-- {
		h = mw[i](h)
	}
	return h
}

func (s *service) basicAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, exists := r.BasicAuth()
			if !exists || user != s.user || pass != s.pass {
				http.Error(w, "user pass unauthenticated",
					http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

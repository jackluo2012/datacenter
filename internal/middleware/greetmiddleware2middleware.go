package middleware

import "net/http"

type GreetMiddleware2Middleware struct {
}

func NewGreetMiddleware2Middleware() *GreetMiddleware2Middleware {
	return &GreetMiddleware2Middleware{}
}

func (m *GreetMiddleware2Middleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO generate middleware implement function, delete after code implementation

		// Passthrough to next handler if need
		next(w, r)
	}
}

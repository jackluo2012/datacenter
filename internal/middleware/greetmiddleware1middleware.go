package middleware

import "net/http"

type GreetMiddleware1Middleware struct {
}

func NewGreetMiddleware1Middleware() *GreetMiddleware1Middleware {
	return &GreetMiddleware1Middleware{}
}

func (m *GreetMiddleware1Middleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO generate middleware implement function, delete after code implementation

		// Passthrough to next handler if need
		next(w, r)
	}
}

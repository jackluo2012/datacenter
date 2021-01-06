package middleware

import (
	"errors"
	"net/http"
)

var (
	AdminError = errors.New("非法的用户请求")
)

type AdminCheckMiddleware struct {
}

func NewAdminCheckMiddleware() *AdminCheckMiddleware {
	return &AdminCheckMiddleware{}
}

func (m *AdminCheckMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// user, pass, _ := r.BasicAuth()
		// logx.Info("user, pass", user, pass)
		// if user == "xxx" && pass == "xxxxx" {
		// 	next(w, r)
		// } else {
		// 	httpx.Error(w, AdminError)
		// }
		next(w, r)
	}
}

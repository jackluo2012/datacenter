package middleware

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"os"

	"github.com/zeromicro/go-zero/rest/httpx"
)

var (
	AdminError = errors.New("非法的用户请求")
)

type AdminCheckMiddleware struct {
	user string
	pass string
}

// NewAdminCheckMiddleware BasicAuth 校验，凭据从环境变量 ADMIN_USER/ADMIN_PASS 读取；
// 未配置时拒绝所有请求（fail-closed），生产环境必须显式配置。
func NewAdminCheckMiddleware() *AdminCheckMiddleware {
	return &AdminCheckMiddleware{
		user: os.Getenv("ADMIN_USER"),
		pass: os.Getenv("ADMIN_PASS"),
	}
}

func (m *AdminCheckMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if m.user == "" || m.pass == "" {
			httpx.Error(w, AdminError)
			return
		}
		user, pass, ok := r.BasicAuth()
		if !ok ||
			subtle.ConstantTimeCompare([]byte(user), []byte(m.user)) != 1 ||
			subtle.ConstantTimeCompare([]byte(pass), []byte(m.pass)) != 1 {
			httpx.Error(w, AdminError)
			return
		}
		next(w, r)
	}
}

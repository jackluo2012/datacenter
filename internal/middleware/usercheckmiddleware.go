package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/zeromicro/go-zero/rest/httpx"
)

var (
	errorUserInfo = errors.New("用户信息获取失败")
	authDeny      = errors.New("用户信息不一致")
)

const (
	userKey = `x-user-id`
)

type UserCheckMiddleware struct {
}

func NewUserCheckMiddleware() *UserCheckMiddleware {
	return &UserCheckMiddleware{}
}

// Handle 防越权校验：请求头 x-user-id 必须与 JWT claims 中的 uid 一致。
// 需挂在 go-zero jwt 中间件之后（claims 已注入 request context）。
func (m *UserCheckMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := r.Header.Get(userKey)
		jwtUserId := r.Context().Value("uid")

		userInt, err := strconv.ParseInt(userId, 10, 64)
		if err != nil {
			httpx.Error(w, errorUserInfo)
			return
		}

		jwtInt, err := json.Number(fmt.Sprintf("%v", jwtUserId)).Int64()
		if err != nil {
			httpx.Error(w, errorUserInfo)
			return
		}

		if jwtInt != userInt {
			httpx.Error(w, authDeny)
			return
		}

		next(w, r)
	}
}

/**
 *  后台 接口更新
 */
func (m *UserCheckMiddleware) AdminHandle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		next(w, r)
	}
}

package middleware

import (
	"errors"
	"net/http"
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

func (m *UserCheckMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		/*
			userId := r.Header.Get(userKey)
			jwtUserId := r.Context().Value("userId")
			logx.Info("r.Context()========", r.Context())
			logx.Info("uid====", r.Context().Value("uid"))
			userInt, err := json.Number(userId).Int64()
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
		*/

		next(w, r)
	}
}

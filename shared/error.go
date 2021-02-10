package shared

import (
	"errors"
	"net/http"
)

const (
	defaultCode = 1001
)

var (
	ErrNotFound               = errors.New("data not find")
	ErrorUserNotFound         = errors.New("用户不存在")
	ErrorNoRequiredParameters = errors.New("必要参数不能为空")
	ErrorUserOperation        = errors.New("用户正在操作中，请稍后重试")
)

type (
	CodeError struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
)

func ErrorHandler(err error) (int, interface{}) {
	return http.StatusConflict, CodeError{
		Code: -1,
		Msg:  err.Error(),
	}
}

func NewCodeError(code int, msg string) error {
	return &CodeError{Code: code, Msg: msg}
}

func NewDefaultError(msg string) error {
	return NewCodeError(defaultCode, msg)
}

func (e *CodeError) Error() string {
	return e.Msg
}

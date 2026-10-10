package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

// ErrNotFound 通用未找到错误（对齐 votes 服务 model 风格）
var ErrNotFound = sqlx.ErrNotFound

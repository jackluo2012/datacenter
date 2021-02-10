package model

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/tal-tech/go-zero/core/stores/cache"
	"github.com/tal-tech/go-zero/core/stores/sqlc"
	"github.com/tal-tech/go-zero/core/stores/sqlx"
	"github.com/tal-tech/go-zero/core/stringx"
	"github.com/tal-tech/go-zero/tools/goctl/model/sql/builderx"
)

var (
	appQuestionsConvertsFieldNames          = builderx.RawFieldNames(&AppQuestionsConverts{})
	appQuestionsConvertsRows                = strings.Join(appQuestionsConvertsFieldNames, ",")
	appQuestionsConvertsRowsExpectAutoSet   = strings.Join(stringx.Remove(appQuestionsConvertsFieldNames, "`id`", "`create_time`", "`update_time`"), ",")
	appQuestionsConvertsRowsWithPlaceHolder = strings.Join(stringx.Remove(appQuestionsConvertsFieldNames, "`id`", "`create_time`", "`update_time`"), "=?,") + "=?"

	cacheAppQuestionsConvertsIdPrefix = "cache#AppQuestionsConverts#id#"
)

type (
	AppQuestionsConvertsModel interface {
		Insert(data AppQuestionsConverts) (sql.Result, error)
		FindOne(id int64) (*AppQuestionsConverts, error)
		Update(data AppQuestionsConverts) error
		Delete(id int64) error
	}

	defaultAppQuestionsConvertsModel struct {
		sqlc.CachedConn
		table string
	}

	AppQuestionsConverts struct {
		Beid       int64          `db:"beid"`  // 对应的平台
		Ptyid      int64          `db:"ptyid"` // 平台id: 1.微信公众号，2.微信小程序，3.支付宝
		Auid       int64          `db:"auid"`
		Username   string         `db:"username"` // 获奖名
		Phone      sql.NullString `db:"phone"`    // 手机号
		Status     int64          `db:"status"`   // 处理（0/1）
		CreateTime time.Time      `db:"create_time"`
		UpdateTime time.Time      `db:"update_time"`
		Uid        int64          `db:"uid"` // 中台表用户的id
		LotteryId  int64          `db:"lottery_id"`
		Id         int64          `db:"id"`
	}
)

func NewAppQuestionsConvertsModel(conn sqlx.SqlConn, c cache.CacheConf) AppQuestionsConvertsModel {
	return &defaultAppQuestionsConvertsModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      "`app_questions_converts`",
	}
}

func (m *defaultAppQuestionsConvertsModel) Insert(data AppQuestionsConverts) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?)", m.table, appQuestionsConvertsRowsExpectAutoSet)
	ret, err := m.ExecNoCache(query, data.Beid, data.Ptyid, data.Auid, data.Username, data.Phone, data.Status, data.Uid, data.LotteryId)

	return ret, err
}

func (m *defaultAppQuestionsConvertsModel) FindOne(id int64) (*AppQuestionsConverts, error) {
	appQuestionsConvertsIdKey := fmt.Sprintf("%s%v", cacheAppQuestionsConvertsIdPrefix, id)
	var resp AppQuestionsConverts
	err := m.QueryRow(&resp, appQuestionsConvertsIdKey, func(conn sqlx.SqlConn, v interface{}) error {
		query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", appQuestionsConvertsRows, m.table)
		return conn.QueryRow(v, query, id)
	})
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultAppQuestionsConvertsModel) Update(data AppQuestionsConverts) error {
	appQuestionsConvertsIdKey := fmt.Sprintf("%s%v", cacheAppQuestionsConvertsIdPrefix, data.Id)
	_, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, appQuestionsConvertsRowsWithPlaceHolder)
		return conn.Exec(query, data.Beid, data.Ptyid, data.Auid, data.Username, data.Phone, data.Status, data.Uid, data.LotteryId, data.Id)
	}, appQuestionsConvertsIdKey)
	return err
}

func (m *defaultAppQuestionsConvertsModel) Delete(id int64) error {

	appQuestionsConvertsIdKey := fmt.Sprintf("%s%v", cacheAppQuestionsConvertsIdPrefix, id)
	_, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
		return conn.Exec(query, id)
	}, appQuestionsConvertsIdKey)
	return err
}

func (m *defaultAppQuestionsConvertsModel) formatPrimary(primary interface{}) string {
	return fmt.Sprintf("%s%v", cacheAppQuestionsConvertsIdPrefix, primary)
}

func (m *defaultAppQuestionsConvertsModel) queryPrimary(conn sqlx.SqlConn, v, primary interface{}) error {
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", appQuestionsConvertsRows, m.table)
	return conn.QueryRow(v, query, primary)
}

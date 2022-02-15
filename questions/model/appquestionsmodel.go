package model

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/builder"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/stringx"
)

var (
	appQuestionsFieldNames          = builder.RawFieldNames(&AppQuestions{})
	appQuestionsRows                = strings.Join(appQuestionsFieldNames, ",")
	appQuestionsRowsExpectAutoSet   = strings.Join(stringx.Remove(appQuestionsFieldNames, "`id`", "`create_time`", "`update_time`"), ",")
	appQuestionsRowsWithPlaceHolder = strings.Join(stringx.Remove(appQuestionsFieldNames, "`id`", "`create_time`", "`update_time`"), "=?,") + "=?"

	cacheAppQuestionsIdPrefix       = "cache#AppQuestions#id#"
	cacheAppQuestionsQuestionPrefix = "cache#AppQuestions#question#"
)

type (
	AppQuestionsModel interface {
		Insert(data AppQuestions) (sql.Result, error)
		FindOne(id int64) (*AppQuestions, error)
		Find(id int64) ([]AppQuestions, error)
		FindOneByQuestion(question string) (*AppQuestions, error)
		Update(data AppQuestions) error
		Delete(id int64) error
	}

	defaultAppQuestionsModel struct {
		sqlc.CachedConn
		table string
	}

	AppQuestions struct {
		Id         int64        `db:"id"`
		Beid       int64        `db:"beid"`        // 对应的平台
		Ptyid      int64        `db:"ptyid"`       // 平台id: 1.微信公众号，2.微信小程序，3.支付宝
		ActivityId int64        `db:"activity_id"` // 活动的id
		Options    string       `db:"options"`     // 选项
		Corrent    string       `db:"corrent"`     // 正确选项（ABCD）
		Status     int64        `db:"status"`      // 状态（01）
		CreatedAt  sql.NullTime `db:"created_at"`
		UpdatedAt  sql.NullTime `db:"updated_at"`
		Question   string       `db:"question"` // 问题
	}
)

func NewAppQuestionsModel(conn sqlx.SqlConn, c cache.CacheConf) AppQuestionsModel {
	return &defaultAppQuestionsModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      "`app_questions`",
	}
}

func (m *defaultAppQuestionsModel) Insert(data AppQuestions) (sql.Result, error) {
	appQuestionsQuestionKey := fmt.Sprintf("%s%v", cacheAppQuestionsQuestionPrefix, data.Question)
	ret, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, appQuestionsRowsExpectAutoSet)
		return conn.Exec(query, data.Beid, data.Ptyid, data.ActivityId, data.Options, data.Corrent, data.Status, data.CreatedAt, data.UpdatedAt, data.Question)
	}, appQuestionsQuestionKey)
	return ret, err
}

func (m *defaultAppQuestionsModel) Find(id int64) ([]AppQuestions, error) {

	var resp []AppQuestions
	query := fmt.Sprintf("select %s from %s where `activity_id` = ?", appQuestionsRows, m.table)
	err := m.QueryRowsNoCache(&resp, query, id)
	//缓存
	return resp, err
}

func (m *defaultAppQuestionsModel) FindOne(id int64) (*AppQuestions, error) {
	appQuestionsIdKey := fmt.Sprintf("%s%v", cacheAppQuestionsIdPrefix, id)
	var resp AppQuestions
	err := m.QueryRow(&resp, appQuestionsIdKey, func(conn sqlx.SqlConn, v interface{}) error {
		query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", appQuestionsRows, m.table)
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

func (m *defaultAppQuestionsModel) FindOneByQuestion(question string) (*AppQuestions, error) {
	appQuestionsQuestionKey := fmt.Sprintf("%s%v", cacheAppQuestionsQuestionPrefix, question)
	var resp AppQuestions
	err := m.QueryRowIndex(&resp, appQuestionsQuestionKey, m.formatPrimary, func(conn sqlx.SqlConn, v interface{}) (i interface{}, e error) {
		query := fmt.Sprintf("select %s from %s where `question` = ? limit 1", appQuestionsRows, m.table)
		if err := conn.QueryRow(&resp, query, question); err != nil {
			return nil, err
		}
		return resp.Id, nil
	}, m.queryPrimary)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultAppQuestionsModel) Update(data AppQuestions) error {
	appQuestionsIdKey := fmt.Sprintf("%s%v", cacheAppQuestionsIdPrefix, data.Id)
	_, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, appQuestionsRowsWithPlaceHolder)
		return conn.Exec(query, data.Beid, data.Ptyid, data.ActivityId, data.Options, data.Corrent, data.Status, data.CreatedAt, data.UpdatedAt, data.Question, data.Id)
	}, appQuestionsIdKey)
	return err
}

func (m *defaultAppQuestionsModel) Delete(id int64) error {
	data, err := m.FindOne(id)
	if err != nil {
		return err
	}

	appQuestionsQuestionKey := fmt.Sprintf("%s%v", cacheAppQuestionsQuestionPrefix, data.Question)
	appQuestionsIdKey := fmt.Sprintf("%s%v", cacheAppQuestionsIdPrefix, id)
	_, err = m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
		return conn.Exec(query, id)
	}, appQuestionsQuestionKey, appQuestionsIdKey)
	return err
}

func (m *defaultAppQuestionsModel) formatPrimary(primary interface{}) string {
	return fmt.Sprintf("%s%v", cacheAppQuestionsIdPrefix, primary)
}

func (m *defaultAppQuestionsModel) queryPrimary(conn sqlx.SqlConn, v, primary interface{}) error {
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", appQuestionsRows, m.table)
	return conn.QueryRow(v, query, primary)
}

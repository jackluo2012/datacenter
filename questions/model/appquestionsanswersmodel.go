package model

import (
	"database/sql"
	"datacenter/questions/rpc/questions"
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
	appQuestionsAnswersFieldNames          = builderx.RawFieldNames(&AppQuestionsAnswers{})
	appQuestionsAnswersRows                = strings.Join(appQuestionsAnswersFieldNames, ",")
	appQuestionsAnswersRowsExpectAutoSet   = strings.Join(stringx.Remove(appQuestionsAnswersFieldNames, "`id`", "`create_time`", "`update_time`"), ",")
	appQuestionsAnswersRowsWithPlaceHolder = strings.Join(stringx.Remove(appQuestionsAnswersFieldNames, "`id`", "`create_time`", "`update_time`"), "=?,") + "=?"

	cacheAppQuestionsAnswersIdPrefix = "cache#AppQuestionsAnswers#id#"
)

type (
	AppQuestionsAnswersModel interface {
		Insert(data AppQuestionsAnswers) (sql.Result, error)
		FindOne(in *questions.GradeReq) (*AppQuestionsAnswers, error)
		Update(data AppQuestionsAnswers) error
		Delete(id int64) error
	}

	defaultAppQuestionsAnswersModel struct {
		sqlc.CachedConn
		table string
	}

	AppQuestionsAnswers struct {
		Id         int64     `db:"id"`
		Beid       int64     `db:"beid"`    // 对应的平台
		Ptyid      int64     `db:"ptyid"`   // 平台id: 1.微信公众号，2.微信小程序，3.支付宝
		Answers    string    `db:"answers"` // 用户答题对错
		Score      string    `db:"score"`   // 得分
		CreateTime time.Time `db:"create_time"`
		UpdateTime time.Time `db:"update_time"`
		Uid        int64     `db:"uid"`  // 中台表用户的id
		Auid       int64     `db:"auid"` // 中台表用户的id
		TestId     int64     `db:"test_id"`
		ActivityId int64     `db:"activity_id"`
	}
)

func NewAppQuestionsAnswersModel(conn sqlx.SqlConn, c cache.CacheConf) AppQuestionsAnswersModel {
	return &defaultAppQuestionsAnswersModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      "`app_questions_answers`",
	}
}

func (m *defaultAppQuestionsAnswersModel) Insert(data AppQuestionsAnswers) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?,?, ?, ?, ?, ?)", m.table, appQuestionsAnswersRowsExpectAutoSet)
	ret, err := m.ExecNoCache(query, data.Beid, data.Ptyid, data.Answers, data.Score, data.Uid, data.Auid, data.TestId, data.ActivityId)

	return ret, err
}

func (m *defaultAppQuestionsAnswersModel) FindOne(in *questions.GradeReq) (*AppQuestionsAnswers, error) {
	appQuestionsAnswersIdKey := fmt.Sprintf("%s%v%v", cacheAppQuestionsAnswersIdPrefix, in.Actid, in.Uid)
	var resp AppQuestionsAnswers
	err := m.QueryRow(&resp, appQuestionsAnswersIdKey, func(conn sqlx.SqlConn, v interface{}) error {
		query := fmt.Sprintf("select %s from %s where activity_id = ? AND `uid` = ? limit 1", appQuestionsAnswersRows, m.table)
		return conn.QueryRow(v, query, in.Actid, in.Uid)
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

func (m *defaultAppQuestionsAnswersModel) Update(data AppQuestionsAnswers) error {
	appQuestionsAnswersIdKey := fmt.Sprintf("%s%v", cacheAppQuestionsAnswersIdPrefix, data.Id)
	_, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, appQuestionsAnswersRowsWithPlaceHolder)
		return conn.Exec(query, data.Beid, data.Ptyid, data.Answers, data.Score, data.Uid, data.TestId, data.ActivityId, data.Id)
	}, appQuestionsAnswersIdKey)
	return err
}

func (m *defaultAppQuestionsAnswersModel) Delete(id int64) error {

	appQuestionsAnswersIdKey := fmt.Sprintf("%s%v", cacheAppQuestionsAnswersIdPrefix, id)
	_, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
		return conn.Exec(query, id)
	}, appQuestionsAnswersIdKey)
	return err
}

func (m *defaultAppQuestionsAnswersModel) formatPrimary(primary interface{}) string {
	return fmt.Sprintf("%s%v", cacheAppQuestionsAnswersIdPrefix, primary)
}

func (m *defaultAppQuestionsAnswersModel) queryPrimary(conn sqlx.SqlConn, v, primary interface{}) error {
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", appQuestionsAnswersRows, m.table)
	return conn.QueryRow(v, query, primary)
}

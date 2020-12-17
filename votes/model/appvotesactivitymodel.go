package model

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/tal-tech/go-zero/core/stores/cache"
	"github.com/tal-tech/go-zero/core/stores/sqlc"
	"github.com/tal-tech/go-zero/core/stores/sqlx"
	"github.com/tal-tech/go-zero/core/stringx"
	"github.com/tal-tech/go-zero/tools/goctl/model/sql/builderx"
)

var (
	appVotesActivityFieldNames          = builderx.FieldNames(&AppVotesActivity{})
	appVotesActivityRows                = strings.Join(appVotesActivityFieldNames, ",")
	appVotesActivityRowsExpectAutoSet   = strings.Join(stringx.Remove(appVotesActivityFieldNames, "actid", "create_time", "update_time"), ",")
	appVotesActivityRowsWithPlaceHolder = strings.Join(stringx.Remove(appVotesActivityFieldNames, "actid", "create_time", "update_time"), "=?,") + "=?"

	cacheAppVotesActivityActidPrefix = "cache#AppVotesActivity#actid#"
)

type (
	AppVotesActivityModel interface {
		Insert(data AppVotesActivity) (sql.Result, error)
		FindOne(actid int64) (*AppVotesActivity, error)
		Update(data AppVotesActivity) error
		Delete(actid int64) error
		IncrVotes(actid int64) error
		IncrView(actid int64) error
		IncrEnroll(actid int64) error
	}

	defaultAppVotesActivityModel struct {
		sqlc.CachedConn
		table string
	}

	AppVotesActivity struct {
		Actid int64 `db:"actid"` // 投票活动的id
		Beid  int64 `db:"beid"`  // 对应的平台
		//		CreateTime time.Time `db:"create_time"`
		Descr      string `db:"descr"`       // 投票活动描述
		EndDate    int64  `db:"end_date"`    // 投票活动结束时间
		EnrollDate int64  `db:"enroll_date"` // 开始投票时间
		Num        int64  `db:"num"`         // 单位
		Ptyid      int64  `db:"ptyid"`       // 平台id: 1.微信公众号，2.微信小程序，3.支付宝
		StartDate  int64  `db:"start_date"`  // 投票活动开始时间
		Status     int64  `db:"status"`      // 0无效，1.是有效
		Title      string `db:"title"`       // 投票活动名称
		Type       int64  `db:"type"`        // 投票的方式:1一次性，2.按天来
		//		UpdateTime time.Time `db:"update_time"`
		Viewcount   int64 `db:"viewcount"`   // 投票活动的总浏览量
		Votecount   int64 `db:"votecount"`   // 投票活动的总票数
		Enrollcount int64 `db:"enrollcount"` // 报名人数
	}
)

func NewAppVotesActivityModel(conn sqlx.SqlConn, c cache.CacheConf) AppVotesActivityModel {
	return &defaultAppVotesActivityModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      "app_votes_activity",
	}
}

// 投票数 +1
func (m *defaultAppVotesActivityModel) IncrVotes(actid int64) error {
	appVotesActivityActidKey := fmt.Sprintf("%s%v", cacheAppVotesActivityActidPrefix, actid)
	_, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set votecount=votecount+1 where actid = ?", m.table)
		return conn.Exec(query, actid)
	}, appVotesActivityActidKey)
	return err
}

//浏览量 +1
func (m *defaultAppVotesActivityModel) IncrView(actid int64) error {
	appVotesActivityActidKey := fmt.Sprintf("%s%v", cacheAppVotesActivityActidPrefix, actid)
	_, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set viewcount=viewcount+1 where actid = ?", m.table)
		return conn.Exec(query, actid)
	}, appVotesActivityActidKey)
	return err
}

//报名数 +1
func (m *defaultAppVotesActivityModel) IncrEnroll(actid int64) error {
	appVotesActivityActidKey := fmt.Sprintf("%s%v", cacheAppVotesActivityActidPrefix, actid)
	_, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set enrollcount=enrollcount+1 where actid = ?", m.table)
		return conn.Exec(query, actid)
	}, appVotesActivityActidKey)
	return err
}

func (m *defaultAppVotesActivityModel) Insert(data AppVotesActivity) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, appVotesActivityRowsExpectAutoSet)
	ret, err := m.ExecNoCache(query, data.Beid, data.Descr, data.EndDate, data.EnrollDate, data.Num, data.Ptyid, data.StartDate, data.Status, data.Title, data.Type, data.Viewcount, data.Votecount)

	return ret, err
}

func (m *defaultAppVotesActivityModel) FindOne(actid int64) (*AppVotesActivity, error) {
	appVotesActivityActidKey := fmt.Sprintf("%s%v", cacheAppVotesActivityActidPrefix, actid)
	var resp AppVotesActivity
	err := m.QueryRow(&resp, appVotesActivityActidKey, func(conn sqlx.SqlConn, v interface{}) error {
		query := fmt.Sprintf("select %s from %s where actid = ? limit 1", appVotesActivityRows, m.table)
		return conn.QueryRow(v, query, actid)
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

func (m *defaultAppVotesActivityModel) Update(data AppVotesActivity) error {
	appVotesActivityActidKey := fmt.Sprintf("%s%v", cacheAppVotesActivityActidPrefix, data.Actid)
	_, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where actid = ?", m.table, appVotesActivityRowsWithPlaceHolder)
		return conn.Exec(query, data.Beid, data.Descr, data.EndDate, data.EnrollDate, data.Num, data.Ptyid, data.StartDate, data.Status, data.Title, data.Type, data.Viewcount, data.Votecount, data.Actid)
	}, appVotesActivityActidKey)
	return err
}

func (m *defaultAppVotesActivityModel) Delete(actid int64) error {

	appVotesActivityActidKey := fmt.Sprintf("%s%v", cacheAppVotesActivityActidPrefix, actid)
	_, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("delete from %s where actid = ?", m.table)
		return conn.Exec(query, actid)
	}, appVotesActivityActidKey)
	return err
}

func (m *defaultAppVotesActivityModel) formatPrimary(primary interface{}) string {
	return fmt.Sprintf("%s%v", cacheAppVotesActivityActidPrefix, primary)
}

func (m *defaultAppVotesActivityModel) queryPrimary(conn sqlx.SqlConn, v, primary interface{}) error {
	query := fmt.Sprintf("select %s from %s where actid = ? limit 1", appVotesActivityRows, m.table)
	return conn.QueryRow(v, query, primary)
}

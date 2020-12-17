package model

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/tal-tech/go-zero/core/stores/sqlc"
	"github.com/tal-tech/go-zero/core/stores/sqlx"
	"github.com/tal-tech/go-zero/core/stringx"
	"github.com/tal-tech/go-zero/tools/goctl/model/sql/builderx"
)

var (
	appVotesFieldNames          = builderx.FieldNames(&AppVotes{})
	appVotesRows                = strings.Join(appVotesFieldNames, ",")
	appVotesRowsExpectAutoSet   = strings.Join(stringx.Remove(appVotesFieldNames, "avid", "create_time", "update_time"), ",")
	appVotesRowsWithPlaceHolder = strings.Join(stringx.Remove(appVotesFieldNames, "avid", "create_time", "update_time"), "=?,") + "=?"
)

type (
	AppVotesModel interface {
		Insert(data AppVotes) (sql.Result, error)
		FindOne(avid int64) (*AppVotes, error)
		Update(data AppVotes) error
		Delete(avid int64) error
		FindByNumWithVotes(aeid, uid, auid int64) (int64, error)
		FindByDaysWithVotes(actid, uid, auid int64, date string) (int64, error)
	}

	defaultAppVotesModel struct {
		conn  sqlx.SqlConn
		table string
	}

	AppVotes struct {
		Actid      int64     `db:"actid"` // 投票活动的id
		Aeid       int64     `db:"aeid"`  // 投票的id
		Auid       int64     `db:"auid"`  // 中台表appuser的id
		Beid       int64     `db:"beid"`  // 对应的平台
		CreateTime time.Time `db:"create_time"`
		Ip         string    `db:"ip"`    // 投票人IP
		Ptyid      int64     `db:"ptyid"` // 平台id: 1.微信公众号，2.微信小程序，3.支付宝
		Uid        int64     `db:"uid"`   // 中台表用户的id
		UpdateTime time.Time `db:"update_time"`
		Avid       int64     `db:"avid"` // 投票人序号：自增
	}
)

func NewAppVotesModel(conn sqlx.SqlConn) AppVotesModel {
	return &defaultAppVotesModel{
		conn:  conn,
		table: "app_votes",
	}
}

/**
 * 获取 按天数投票
 * aeid 活动的id
 * uid 用户的id
 * auid snsid
 * date
 */

func (m *defaultAppVotesModel) FindByDaysWithVotes(actid, uid, auid int64, date string) (int64, error) {
	query := fmt.Sprintf("select count(*) as count from %s where actid = ? AND uid = ? AND auid = ? AND create_time like ? limit 1", m.table)
	var count int64
	err := m.conn.QueryRowPartial(&count, query, actid, uid, auid, date+"%")
	if err != nil {
		return 0, err
	}
	return count, nil
}

/**
 * 获取 按次数投票
 * aeid 活动的id
 * uid 用户的id
 * auid snsid
 */
func (m *defaultAppVotesModel) FindByNumWithVotes(actid, uid, auid int64) (int64, error) {
	query := fmt.Sprintf("select count(*) as count from %s where actid = ? AND uid = ? AND auid = ? limit 1", appVotesRows, m.table)
	var count int64
	err := m.conn.QueryRowPartial(&count, query, actid, uid, auid)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m *defaultAppVotesModel) Insert(data AppVotes) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?,  ?, ?)", m.table, appVotesRowsExpectAutoSet)
	ret, err := m.conn.Exec(query, data.Actid, data.Aeid, data.Auid, data.Beid, data.Ip, data.Ptyid, data.Uid)
	return ret, err
}

func (m *defaultAppVotesModel) FindOne(avid int64) (*AppVotes, error) {
	query := fmt.Sprintf("select %s from %s where avid = ? limit 1", appVotesRows, m.table)
	var resp AppVotes
	err := m.conn.QueryRow(&resp, query, avid)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultAppVotesModel) Update(data AppVotes) error {
	query := fmt.Sprintf("update %s set %s where avid = ?", m.table, appVotesRowsWithPlaceHolder)
	_, err := m.conn.Exec(query, data.Actid, data.Aeid, data.Auid, data.Beid, data.Ip, data.Ptyid, data.Uid, data.Avid)
	return err
}

func (m *defaultAppVotesModel) Delete(avid int64) error {
	query := fmt.Sprintf("delete from %s where avid = ?", m.table)
	_, err := m.conn.Exec(query, avid)
	return err
}

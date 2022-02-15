package model

import (
	"database/sql"
	"datacenter/votes/rpc/votes"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/builder"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/stringx"
)

var (
	appEnrollFieldNames          = builder.RawFieldNames(&AppEnroll{})
	appEnrollRows                = strings.Join(appEnrollFieldNames, ",")
	appEnrollRowsExpectAutoSet   = strings.Join(stringx.Remove(appEnrollFieldNames, "aeid", "create_time", "update_time"), ",")
	appEnrollRowsWithPlaceHolder = strings.Join(stringx.Remove(appEnrollFieldNames, "aeid", "create_time", "update_time"), "=?,") + "=?"

	cacheAppEnrollAeidPrefix       = "cache#AppEnroll#aeid#"
	cacheAppEnrollActUidAuidPrefix = "cache#AppEnroll#actid#uid#auid#"
)

type (
	AppEnrollModel interface {
		Insert(data AppEnroll) (sql.Result, error)
		FindOne(aeid int64) (*AppEnroll, error)
		Update(data AppEnroll) error
		Delete(aeid int64) error
		IncrView(aeid int64) error
		IncrVotes(aeid int64) error
		GetActIdOrAcidExist(actid, uid, auid int64) (int64, error)
		Find(req *votes.ActidReq) ([]AppEnroll, error)
	}

	defaultAppEnrollModel struct {
		sqlc.CachedConn
		table string
	}

	AppEnroll struct {
		Actid      int64     `db:"actid"`       // 投票活动的id
		Address    string    `db:"address"`     // 地址
		Auid       int64     `db:"auid"`        // 中台表appuser的id
		Beid       int64     `db:"beid"`        // 对应的平台
		CreateTime time.Time `db:"create_time"` // 创建时间
		Descr      string    `db:"descr"`       // 介绍
		Images     string    `db:"images"`      // 图片
		Name       string    `db:"name"`        // 名字
		Ptyid      int64     `db:"ptyid"`       // 平台id: 1.微信公众号，2.微信小程序，3.支付宝
		Status     int64     `db:"status"`      // 0,未审核，1.审核通过，2.删除
		Uid        int64     `db:"uid"`         // 中台表用户的id
		UpdateTime time.Time `db:"update_time"`
		Viewcount  int64     `db:"viewcount"` // 浏览数
		Votecount  int64     `db:"votecount"` // 投票数
		Aeid       int64     `db:"aeid"`
	}
)

func NewAppEnrollModel(conn sqlx.SqlConn, c cache.CacheConf) AppEnrollModel {
	return &defaultAppEnrollModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      "app_enroll",
	}
}

func (m *defaultAppEnrollModel) Find(req *votes.ActidReq) ([]AppEnroll, error) {
	var resp []AppEnroll
	var err error
	if req.Keyword == "" {
		query := fmt.Sprintf("select %s from %s where actid = ? AND status=1 Order By votecount desc Limit ?,?", appEnrollRows, m.table)
		err = m.QueryRowsNoCache(&resp, query, req.Actid, req.Limit.Offset, req.Limit.Size)
	} else {
		query := fmt.Sprintf("select %s from %s where actid = ? AND status=1 AND (aeid=? OR name like ?) Order By votecount desc Limit ?,?", appEnrollRows, m.table)
		err = m.QueryRowsNoCache(&resp, query, req.Actid, req.Keyword, "%"+req.Keyword+"%", req.Limit.Offset, req.Limit.Size)
	}
	return resp, err
}

func (m *defaultAppEnrollModel) Insert(data AppEnroll) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, appEnrollRowsExpectAutoSet)
	ret, err := m.ExecNoCache(query, data.Actid, data.Address, data.Auid, data.Beid, data.Descr, data.Images, data.Name, data.Ptyid, data.Status, data.Uid, data.Viewcount, data.Votecount)
	//添加的时候删除缓存
	appEnrollAeidKey := fmt.Sprintf("%s%v%v%v", cacheAppEnrollActUidAuidPrefix, data.Actid, data.Uid, data.Auid)
	m.DelCache(appEnrollAeidKey)
	return ret, err
}

func (m *defaultAppEnrollModel) FindOne(aeid int64) (*AppEnroll, error) {
	appEnrollAeidKey := fmt.Sprintf("%s%v", cacheAppEnrollAeidPrefix, aeid)
	var resp AppEnroll
	err := m.QueryRow(&resp, appEnrollAeidKey, func(conn sqlx.SqlConn, v interface{}) error {
		query := fmt.Sprintf("select %s from %s where aeid = ? limit 1", appEnrollRows, m.table)
		return conn.QueryRow(v, query, aeid)
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

func (m *defaultAppEnrollModel) Update(data AppEnroll) error {
	appEnrollAeidKey := fmt.Sprintf("%s%v", cacheAppEnrollAeidPrefix, data.Aeid)
	_, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where aeid = ?", m.table, appEnrollRowsWithPlaceHolder)
		return conn.Exec(query, data.Actid, data.Address, data.Auid, data.Beid, data.Descr, data.Images, data.Name, data.Ptyid, data.Status, data.Uid, data.Viewcount, data.Votecount, data.Aeid)
	}, appEnrollAeidKey)
	return err
}

func (m *defaultAppEnrollModel) Delete(aeid int64) error {

	appEnrollAeidKey := fmt.Sprintf("%s%v", cacheAppEnrollAeidPrefix, aeid)
	_, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("delete from %s where aeid = ?", m.table)
		return conn.Exec(query, aeid)
	}, appEnrollAeidKey)
	return err
}

func (m *defaultAppEnrollModel) formatPrimary(primary interface{}) string {
	return fmt.Sprintf("%s%v", cacheAppEnrollAeidPrefix, primary)
}

func (m *defaultAppEnrollModel) queryPrimary(conn sqlx.SqlConn, v, primary interface{}) error {
	query := fmt.Sprintf("select %s from %s where aeid = ? limit 1", appEnrollRows, m.table)
	return conn.QueryRow(v, query, primary)
}

// 投票数 +1
func (m *defaultAppEnrollModel) IncrVotes(aeid int64) error {
	appEnrollAeidKey := fmt.Sprintf("%s%v", cacheAppEnrollAeidPrefix, aeid)
	_, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set votecount=votecount+1 where aeid = ?", m.table)
		return conn.Exec(query, aeid)
	}, appEnrollAeidKey)
	return err
}

//浏览量 +1
func (m *defaultAppEnrollModel) IncrView(aeid int64) error {
	appEnrollAeidKey := fmt.Sprintf("%s%v", cacheAppEnrollAeidPrefix, aeid)
	_, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set viewcount=viewcount+1 where aeid = ?", m.table)
		return conn.Exec(query, aeid)
	}, appEnrollAeidKey)
	return err
}

// 获取 用户是否重复报名
func (m *defaultAppEnrollModel) GetActIdOrAcidExist(actid, uid, auid int64) (int64, error) {
	appEnrollAeidKey := fmt.Sprintf("%s%v%v%v", cacheAppEnrollActUidAuidPrefix, actid, uid, auid)
	var count int64
	err := m.QueryRow(&count, appEnrollAeidKey, func(conn sqlx.SqlConn, v interface{}) error {
		query := fmt.Sprintf("select count(*) as count from %s where actid = ? AND uid = ? AND auid = ? limit 1", m.table)
		return conn.QueryRowPartial(v, query, actid, uid, auid)
	})
	return count, err
}

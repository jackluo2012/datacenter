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
	appUserFieldNames          = builderx.FieldNames(&AppUser{})
	appUserRows                = strings.Join(appUserFieldNames, ",")
	appUserRowsExpectAutoSet   = strings.Join(stringx.Remove(appUserFieldNames, "auid", "create_time", "update_time"), ",")
	appUserRowsWithPlaceHolder = strings.Join(stringx.Remove(appUserFieldNames, "auid", "create_time", "update_time"), "=?,") + "=?"

	cacheAppUserAuidPrefix            = "cache#AppUser#auid#"
	cacheAppUserBeidPtyidOpenidPrefix = "cache#AppUser#beid#ptyid#openid#"
)

type (
	AppUserModel struct {
		sqlc.CachedConn
		table string
	}

	AppUser struct {
		Auid       int64     `db:"auid"`
		Avator     string    `db:"avator"` // 头像
		Beid       int64     `db:"beid"`   // 对应的平台
		City       string    `db:"city"`
		Country    string    `db:"country"`
		CreateTime time.Time `db:"create_time"` // 创建时间
		Nickname   string    `db:"nickname"`    // 昵称
		Privilege  string    `db:"privilege"`
		Province   string    `db:"province"`
		Ptyid      int64     `db:"ptyid"` // 平台id: 1.微信公众号，2.微信小程序，3.支付宝
		Sex        int64     `db:"sex"`   // 性别
		Uid        int64     `db:"uid"`   // 对应中台表中的id
		UnionId    string    `db:"unionid"`
		UpdateTime time.Time `db:"update_time"`
		Openid     string    `db:"openid"` // 社交属性的openid
	}
)

func NewAppUserModel(conn sqlx.SqlConn, c cache.CacheConf) *AppUserModel {
	return &AppUserModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      "app_user",
	}
}

func (m *AppUserModel) Insert(data AppUser) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?,  ?, ?, ?, ?, ?)", m.table, appUserRowsExpectAutoSet)
	ret, err := m.ExecNoCache(query, data.Avator, data.Beid, data.City, data.Country, data.Nickname, data.Privilege, data.Province, data.Ptyid, data.Sex, data.Uid, data.UnionId, data.Openid)
	//删除缓存
	appUserAuidKey := fmt.Sprintf("%s%v%v%s", cacheAppUserBeidPtyidOpenidPrefix, data.Beid, data.Ptyid, data.Openid)
	m.DelCache(appUserAuidKey)
	return ret, err
}

/**
 *	根据openid 获取 用户信息
 */
func (m *AppUserModel) FindOneByOpenid(beid, ptyid int64, openid string) (*AppUser, error) {
	appUserAuidKey := fmt.Sprintf("%s%v%v%s", cacheAppUserBeidPtyidOpenidPrefix, beid, ptyid, openid)
	var resp AppUser
	err := m.QueryRow(&resp, appUserAuidKey, func(conn sqlx.SqlConn, v interface{}) error {
		query := fmt.Sprintf("select %s from %s where beid=? AND ptyid=? AND openid = ? limit 1", appUserRows, m.table)
		return conn.QueryRow(v, query, beid, ptyid, openid)
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

func (m *AppUserModel) Update(data AppUser) error {
	appUserAuidKey := fmt.Sprintf("%s%v", cacheAppUserAuidPrefix, data.Auid)
	_, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where auid = ?", m.table, appUserRowsWithPlaceHolder)
		return conn.Exec(query, data.Avator, data.Beid, data.City, data.Country, data.Nickname, data.Privilege, data.Province, data.Ptyid, data.Sex, data.Uid, data.UnionId, data.Openid, data.Auid)
	}, appUserAuidKey)
	return err
}

func (m *AppUserModel) Delete(auid int64) error {

	appUserAuidKey := fmt.Sprintf("%s%v", cacheAppUserAuidPrefix, auid)
	_, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("delete from %s where auid = ?", m.table)
		return conn.Exec(query, auid)
	}, appUserAuidKey)
	return err
}

func (m *AppUserModel) formatPrimary(primary interface{}) string {
	return fmt.Sprintf("%s%v", cacheAppUserAuidPrefix, primary)
}

func (m *AppUserModel) queryPrimary(conn sqlx.SqlConn, v, primary interface{}) error {
	query := fmt.Sprintf("select %s from %s where auid = ? limit 1", appUserRows, m.table)
	return conn.QueryRow(v, query, primary)
}

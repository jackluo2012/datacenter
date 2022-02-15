package model

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"errors"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/builder"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/stringx"
)

var (
	appQuestionsLotteriesFieldNames          = builder.RawFieldNames(&AppQuestionsLotteries{})
	appQuestionsLotteriesRows                = strings.Join(appQuestionsLotteriesFieldNames, ",")
	appQuestionsLotteriesRowsExpectAutoSet   = strings.Join(stringx.Remove(appQuestionsLotteriesFieldNames, "`id`", "`create_time`", "`update_time`"), ",")
	appQuestionsLotteriesRowsWithPlaceHolder = strings.Join(stringx.Remove(appQuestionsLotteriesFieldNames, "`id`", "`create_time`", "`update_time`"), "=?,") + "=?"

	cacheAppQuestionsLotteriesIdPrefix = "cache#AppQuestionsLotteries#id#"
)

type (
	AppQuestionsLotteriesModel interface {
		Insert(data AppQuestionsLotteries) (sql.Result, error)
		FindOne(id int64) (*AppQuestionsLotteries, error)
		Update(data AppQuestionsLotteries) error
		Delete(id int64) error
		TurnTable(data AppQuestionsLotteries) error
	}

	defaultAppQuestionsLotteriesModel struct {
		sqlc.CachedConn
		table string
	}

	AppQuestionsLotteries struct {
		Id         int64     `db:"id"`
		Beid       int64     `db:"beid"`  // 对应的平台
		Ptyid      int64     `db:"ptyid"` // 平台id: 1.微信公众号，2.微信小程序，3.支付宝
		Auid       int64     `db:"auid"`
		IsWinning  int64     `db:"is_winning"` // 是否中奖（0/1）
		IsConvert  int64     `db:"is_convert"` // 是否兑奖（0/1）
		CreateTime time.Time `db:"create_time"`
		UpdateTime time.Time `db:"update_time"`
		Uid        int64     `db:"uid"` // 中台表用户的id
		ActivityId int64     `db:"activity_id"`
		AnswerId   int64     `db:"answer_id"`
		AwardId    int64     `db:"award_id"`
	}
)

func NewAppQuestionsLotteriesModel(conn sqlx.SqlConn, c cache.CacheConf) AppQuestionsLotteriesModel {
	return &defaultAppQuestionsLotteriesModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      "`app_questions_lotteries`",
	}
}

func (m *defaultAppQuestionsLotteriesModel) Insert(data AppQuestionsLotteries) (sql.Result, error) {
	query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, appQuestionsLotteriesRowsExpectAutoSet)
	ret, err := m.ExecNoCache(query, data.Beid, data.Ptyid, data.Auid, data.IsWinning, data.IsConvert, data.Uid, data.ActivityId, data.AnswerId, data.AwardId)

	return ret, err
}

func (m *defaultAppQuestionsLotteriesModel) FindOne(id int64) (*AppQuestionsLotteries, error) {
	appQuestionsLotteriesIdKey := fmt.Sprintf("%s%v", cacheAppQuestionsLotteriesIdPrefix, id)
	var resp AppQuestionsLotteries
	err := m.QueryRow(&resp, appQuestionsLotteriesIdKey, func(conn sqlx.SqlConn, v interface{}) error {
		query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", appQuestionsLotteriesRows, m.table)
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

/**
 * 添加中奖记录，并减库存
 */
func (m *defaultAppQuestionsLotteriesModel) TurnTable(data AppQuestionsLotteries) error {
	//组织sql
	insertsql := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, appQuestionsLotteriesRowsExpectAutoSet)
	err := m.CachedConn.Transact(func(session sqlx.Session) error {
		//先查
		var number int64
		err := session.QueryRowPartial(&number, "select count(*) as count from app_questions_awards WHERE activity_id = ? AND id = ?", data.ActivityId, data.AwardId)
		if err != nil {
			return err
		}
		if number <= 0 {
			return errors.New("库存不足")
		}
		//添加到奖品表
		stmt, err := session.Prepare(insertsql)
		if err != nil {
			return err
		}
		defer stmt.Close()
		// 返回任何错误都会回滚事务
		if _, err := stmt.Exec(data.Beid, data.Ptyid, data.Auid, data.IsWinning, data.IsConvert, data.Uid, data.ActivityId, data.AnswerId, data.AwardId); err != nil {
			logx.Errorf("insert Lotteries stmt exec: %s", err)
			return err
		}
		// 修改奖品的数量
		if _, err := session.Exec("update app_questions_awards set number=? where activity_id = ? AND id = ?", number-1, data.ActivityId, data.AwardId); err != nil {
			return err
		}

		// 还可以继续执行 insert/update/delete 相关操作
		return nil
	})
	return err
}

func (m *defaultAppQuestionsLotteriesModel) Update(data AppQuestionsLotteries) error {
	appQuestionsLotteriesIdKey := fmt.Sprintf("%s%v", cacheAppQuestionsLotteriesIdPrefix, data.Id)
	_, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, appQuestionsLotteriesRowsWithPlaceHolder)
		return conn.Exec(query, data.Beid, data.Ptyid, data.Auid, data.IsWinning, data.IsConvert, data.Uid, data.ActivityId, data.AnswerId, data.AwardId, data.Id)
	}, appQuestionsLotteriesIdKey)
	return err
}

func (m *defaultAppQuestionsLotteriesModel) Delete(id int64) error {

	appQuestionsLotteriesIdKey := fmt.Sprintf("%s%v", cacheAppQuestionsLotteriesIdPrefix, id)
	_, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
		return conn.Exec(query, id)
	}, appQuestionsLotteriesIdKey)
	return err
}

func (m *defaultAppQuestionsLotteriesModel) formatPrimary(primary interface{}) string {
	return fmt.Sprintf("%s%v", cacheAppQuestionsLotteriesIdPrefix, primary)
}

func (m *defaultAppQuestionsLotteriesModel) queryPrimary(conn sqlx.SqlConn, v, primary interface{}) error {
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", appQuestionsLotteriesRows, m.table)
	return conn.QueryRow(v, query, primary)
}

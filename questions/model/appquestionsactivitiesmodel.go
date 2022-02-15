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
	appQuestionsActivitiesFieldNames          = builder.RawFieldNames(&AppQuestionsActivities{})
	appQuestionsActivitiesRows                = strings.Join(appQuestionsActivitiesFieldNames, ",")
	appQuestionsActivitiesRowsExpectAutoSet   = strings.Join(stringx.Remove(appQuestionsActivitiesFieldNames, "`id`", "`create_time`", "`update_time`"), ",")
	appQuestionsActivitiesRowsWithPlaceHolder = strings.Join(stringx.Remove(appQuestionsActivitiesFieldNames, "`id`", "`create_time`", "`update_time`"), "=?,") + "=?"

	cacheAppQuestionsActivitiesIdPrefix    = "cache#AppQuestionsActivities#id#"
	cacheAppQuestionsActivitiesTitlePrefix = "cache#AppQuestionsActivities#title#"
)

type (
	AppQuestionsActivitiesModel interface {
		Insert(data AppQuestionsActivities) (sql.Result, error)
		FindOne(id int64) (*AppQuestionsActivities, error)
		FindOneByTitle(title string) (*AppQuestionsActivities, error)
		Update(data AppQuestionsActivities) error
		Delete(id int64) error
	}

	defaultAppQuestionsActivitiesModel struct {
		sqlc.CachedConn
		table string
	}

	AppQuestionsActivities struct {
		Id           int64        `db:"id"`
		Beid         int64        `db:"beid"`          // 对应的平台
		Ptyid        int64        `db:"ptyid"`         // 平台id: 1.微信公众号，2.微信小程序，3.支付宝
		StartDate    int64        `db:"start_date"`    // 投票活动开始时间
		EndDate      int64        `db:"end_date"`      // 投票活动结束时间
		GetScore     float64      `db:"get_score"`     // 及格分数
		ActivityWeek string       `db:"activity_week"` // 活动周期（可多选）,分隔
		Header       string       `db:"header"`        // 分享标题
		Des          string       `db:"des"`           // 分享文字
		Image        string       `db:"image"`         // 分享图片
		Rule         string       `db:"rule"`          // 抽奖规则
		Status       int64        `db:"status"`        // 状态
		CreatedAt    sql.NullTime `db:"created_at"`
		UpdatedAt    sql.NullTime `db:"updated_at"`
		Title        string       `db:"title"` // 活动主题
	}
)

func NewAppQuestionsActivitiesModel(conn sqlx.SqlConn, c cache.CacheConf) AppQuestionsActivitiesModel {
	return &defaultAppQuestionsActivitiesModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      "`app_questions_activities`",
	}
}

func (m *defaultAppQuestionsActivitiesModel) Insert(data AppQuestionsActivities) (sql.Result, error) {
	appQuestionsActivitiesTitleKey := fmt.Sprintf("%s%v", cacheAppQuestionsActivitiesTitlePrefix, data.Title)
	ret, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, appQuestionsActivitiesRowsExpectAutoSet)
		return conn.Exec(query, data.Beid, data.Ptyid, data.StartDate, data.EndDate, data.GetScore, data.ActivityWeek, data.Header, data.Des, data.Image, data.Rule, data.Status, data.CreatedAt, data.UpdatedAt, data.Title)
	}, appQuestionsActivitiesTitleKey)
	return ret, err
}

func (m *defaultAppQuestionsActivitiesModel) FindOne(id int64) (*AppQuestionsActivities, error) {
	appQuestionsActivitiesIdKey := fmt.Sprintf("%s%v", cacheAppQuestionsActivitiesIdPrefix, id)
	var resp AppQuestionsActivities
	err := m.QueryRow(&resp, appQuestionsActivitiesIdKey, func(conn sqlx.SqlConn, v interface{}) error {
		query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", appQuestionsActivitiesRows, m.table)
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

func (m *defaultAppQuestionsActivitiesModel) FindOneByTitle(title string) (*AppQuestionsActivities, error) {
	appQuestionsActivitiesTitleKey := fmt.Sprintf("%s%v", cacheAppQuestionsActivitiesTitlePrefix, title)
	var resp AppQuestionsActivities
	err := m.QueryRowIndex(&resp, appQuestionsActivitiesTitleKey, m.formatPrimary, func(conn sqlx.SqlConn, v interface{}) (i interface{}, e error) {
		query := fmt.Sprintf("select %s from %s where `title` = ? limit 1", appQuestionsActivitiesRows, m.table)
		if err := conn.QueryRow(&resp, query, title); err != nil {
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

func (m *defaultAppQuestionsActivitiesModel) Update(data AppQuestionsActivities) error {
	appQuestionsActivitiesIdKey := fmt.Sprintf("%s%v", cacheAppQuestionsActivitiesIdPrefix, data.Id)
	_, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, appQuestionsActivitiesRowsWithPlaceHolder)
		return conn.Exec(query, data.Beid, data.Ptyid, data.StartDate, data.EndDate, data.GetScore, data.ActivityWeek, data.Header, data.Des, data.Image, data.Rule, data.Status, data.CreatedAt, data.UpdatedAt, data.Title, data.Id)
	}, appQuestionsActivitiesIdKey)
	return err
}

func (m *defaultAppQuestionsActivitiesModel) Delete(id int64) error {
	data, err := m.FindOne(id)
	if err != nil {
		return err
	}

	appQuestionsActivitiesIdKey := fmt.Sprintf("%s%v", cacheAppQuestionsActivitiesIdPrefix, id)
	appQuestionsActivitiesTitleKey := fmt.Sprintf("%s%v", cacheAppQuestionsActivitiesTitlePrefix, data.Title)
	_, err = m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
		return conn.Exec(query, id)
	}, appQuestionsActivitiesIdKey, appQuestionsActivitiesTitleKey)
	return err
}

func (m *defaultAppQuestionsActivitiesModel) formatPrimary(primary interface{}) string {
	return fmt.Sprintf("%s%v", cacheAppQuestionsActivitiesIdPrefix, primary)
}

func (m *defaultAppQuestionsActivitiesModel) queryPrimary(conn sqlx.SqlConn, v, primary interface{}) error {
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", appQuestionsActivitiesRows, m.table)
	return conn.QueryRow(v, query, primary)
}

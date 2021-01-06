package logic

import (
	"context"
	"datacenter/search/rpc/internal/svc"
	"datacenter/search/rpc/search"
	"datacenter/search/rpc/searchclient"
	"datacenter/shared"
	"encoding/json"
	"net/url"
	"sync"

	"github.com/tal-tech/go-zero/core/logx"
)

//数据初始化模板事例
//一些数据结构
type PageData struct {
	ResultMsg string
	Data      Data
}
type Data struct {
	Header    []searchclient.ArticleReq
	Normal    []searchclient.ArticleReq
	TimeStamp int64
	Page      int64
	ViewTimes int64
}

type ArticleInitLogic struct {
	ctx    context.Context
	lock   *sync.Mutex
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewArticleInitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ArticleInitLogic {
	return &ArticleInitLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		lock:   new(sync.Mutex),
		Logger: logx.WithContext(ctx),
	}
}

//这个是个递归
func (l *ArticleInitLogic) ArticleRecursion(gourl string, form url.Values, once bool) error {
	body, err := shared.HttpPostForm(gourl, form)
	if err != nil {
		return err
	}
	var data PageData
	err = json.Unmarshal(body, &data)
	if err != nil {
		return err
	}
	if len(data.Data.Header) > 0 {
		for _, art := range data.Data.Header {
			l.svcCtx.ArticeEs.Index(&art)
		}
	}
	if len(data.Data.Normal) > 0 {
		for _, art := range data.Data.Normal {
			id, err := l.svcCtx.ArticeEs.Index(&art)
			if err != nil {
				logx.Info("err==", err)
			}
			logx.Info("id=", id)
		}
		if !once {
			err = l.ArticleRecursion(gourl, url.Values{"TimeStamp": []string{shared.Int64ToStr(data.Data.TimeStamp)}, "Page": []string{shared.Int64ToStr(data.Data.Page)}, "ViewTimes": []string{shared.Int64ToStr(data.Data.ViewTimes)}}, once)
		}
	}
	return err
}

/**
 * 初始化抓取数据
 */
func (l *ArticleInitLogic) ArticleInit(in *search.Request) (*search.Response, error) {
	//开始锁定
	l.lock.Lock()
	//err := l.ArticleRecursion("http://xxxxxxxxxxx", url.Values{}, in.Once)
	l.lock.Unlock()
	// if err != nil {
	// 	return nil, err
	// }
	return &search.Response{}, nil
}

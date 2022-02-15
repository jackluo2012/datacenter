package dao

import (
	"context"
	"datacenter/search/rpc/searchclient"
	"fmt"
	"reflect"

	"github.com/olivere/elastic/v7"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	author     = "tanzi"
	project    = "article"
	mappingTpl = `	
	{
		"settings":{
			"number_of_shards":1,
			"number_of_replicas":0
		},
		"mappings":{
			"properties" : {
			  "ImageUrl" : {
				"type" : "keyword"
			  },
			  "NewsId" : {
				"type" : "keyword"
			  },
			  "NewsTitle" : {
				"type" : "text",
				"analyzer" : "ik_max_word",
          		"search_analyzer" : "ik_smart",
				"fields" : {
				  "keyword" : {
					"type" : "keyword",
					"ignore_above" : 256
				  }
				}
			  }
			}
		  }
		}
	}`
	esRetryLimit = 3 //bulk 错误重试机制
)

type ArticelES struct {
	index   string
	mapping string
	client  *elastic.Client
}

func NewArticelES(client *elastic.Client) (*ArticelES, error) {
	index := fmt.Sprintf("%s_%s", author, project)
	articelEs := &ArticelES{
		client:  client,
		index:   index,
		mapping: mappingTpl,
	}
	//创建表
	exists, err := articelEs.client.IndexExists(articelEs.index).Do(context.Background())
	if err != nil {
		logx.Info(err)
		return nil, err
	}
	if !exists {
		_, err := articelEs.client.CreateIndex(articelEs.index).Body(articelEs.mapping).Do(context.Background())
		if err != nil {
			logx.Info("userEs init failed err is %s\n", err)
			return nil, err
		}
	}
	return articelEs, nil
}

//索引
func (aes *ArticelES) Index(art *searchclient.ArticleReq) (string, error) {
	res, err := aes.client.Index().Index(aes.index).Id(art.NewsId).BodyJson(art).Do(context.Background())
	if err != nil {
		return "", err
	}
	return res.Id, err
}

//得到值

func (aes *ArticelES) Search(search *searchclient.SearchReq) ([]*searchclient.ArticleReq, error) {
	searcharticleLists := make([]*searchclient.ArticleReq, 0)
	//termQuery := elastic.NewTermQuery("NewsTitle", search.Keyword)
	matchQuery := elastic.NewMatchQuery("NewsTitle", search.Keyword)

	searchResult, err := aes.client.Search().
		Index(aes.index). // search in index "twitter"
		//Source("NewsId,NewsTitle").
		//Source([]string{"NewsId", "NewsTitle"}).
		Query(matchQuery). // specify the query
		Sort("NewsId", false).
		From(int(search.Limit.Offset)).
		Size(int(search.Limit.Size)). // take documents 0-9
		Pretty(true).                 // pretty print request and response JSON
		Do(context.Background())      // execute

	if err != nil {
		return searcharticleLists, err
	}
	var articel searchclient.ArticleReq
	for _, item := range searchResult.Each(reflect.TypeOf(articel)) {
		t, ok := item.(searchclient.ArticleReq)
		if ok {
			searcharticleLists = append(searcharticleLists, &t)
		}
	}

	return searcharticleLists, nil
}

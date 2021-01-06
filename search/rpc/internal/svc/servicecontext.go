package svc

import (
	"datacenter/search/rpc/dao"
	"datacenter/search/rpc/internal/config"
	"strings"

	"github.com/olivere/elastic/v7"
)

type ServiceContext struct {
	c        config.Config
	Esconn   *elastic.Client
	ArticeEs *dao.ArticelES
}

func NewServiceContext(c config.Config) *ServiceContext {

	client, err := elastic.NewClient(elastic.SetSniff(false), elastic.SetURL(strings.Join(c.Esconfig.Urls, ",")), elastic.SetBasicAuth(c.Esconfig.User, c.Esconfig.Password))
	if err != nil {
		panic(err)
	}
	ArticeEs, err := dao.NewArticelES(client)
	if err != nil {
		panic(err)
	}
	return &ServiceContext{
		c:        c,
		Esconn:   client,
		ArticeEs: ArticeEs,
	}
}

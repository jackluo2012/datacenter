package logic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"datacenter/shared"
	"datacenter/votes/model"
	"datacenter/votes/rpc/internal/svc"
	"datacenter/votes/rpc/votes"

	"github.com/tal-tech/go-zero/core/logx"
	"github.com/tal-tech/go-zero/core/stores/redis"
)

type VotesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewVotesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VotesLogic {
	return &VotesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *VotesLogic) Votes(in *votes.VotesReq) (*votes.VotesResp, error) {
	//获取 活动信息
	//加锁
	redisLock := redis.NewRedisLock(l.svcCtx.RedisConn, shared.GetVoteUidAuidLockKey(in.Uid, in.Auid))
	// 2. 可选操作，设置 redislock 过期时间
	redisLock.SetExpire(shared.VotesUidAuidLockExpire)
	if ok, err := redisLock.Acquire(); !ok || err != nil {
		return nil, VotesLock
	}
	// 3. 释放锁
	defer redisLock.Release()
	// 获取 投票活动的信息
	activit, err := l.svcCtx.AppVotesActivityModel.FindOne(in.Actid)
	if err != nil {
		return nil, ActivityDoesNotExist
	}
	// 检查 活动的状态
	if activit.Status != 1 {
		return nil, ActivityDoesNotOpen
	}
	//检查 活动是否可以投票
	nowtime := time.Now().Unix()
	if activit.StartDate > nowtime {
		return nil, ActivityDoesNotEnroll
	}
	//活动结束
	if activit.EndDate < nowtime {
		return nil, ActivityEnd
	}
	//判断投票方式 1.按次数投票，2.按天数投票
	if activit.Type == 2 {
		count, err := l.svcCtx.AppVotesModel.FindByDaysWithVotes(in.Actid, in.Uid, in.Auid, shared.GetDate())
		if err != nil {
			return nil, VotesFailt
		}
		if count >= activit.Num {
			return nil, errors.New(fmt.Sprintf("每天只能投%d票", activit.Num))
		}
	} else {
		count, err := l.svcCtx.AppVotesModel.FindByNumWithVotes(in.Actid, in.Uid, in.Auid)
		if err != nil {
			return nil, VotesFailt
		}
		if count >= activit.Num {
			return nil, errors.New(fmt.Sprintf("每人只能投%d票", activit.Num))
		}
	}
	//检查 作品的状态
	appenroll, err := l.svcCtx.AppEnrollModel.FindOne(in.Aeid)
	if err != nil {
		return nil, EnrollNotExist
	}
	if appenroll.Status != 1 {
		return nil, EnrollNotExist
	}
	//写入投票表
	_, err = l.svcCtx.AppVotesModel.Insert(model.AppVotes{
		Actid: in.Actid,
		Ptyid: in.Ptyid,
		Aeid:  in.Aeid,
		Uid:   in.Uid,
		Auid:  in.Auid,
		Beid:  in.Beid,
		Ip:    in.Ip,
	})
	if err != nil {
		return nil, err
	}
	// 更新投票和浏览
	l.svcCtx.AppEnrollModel.IncrVotes(in.Aeid)
	l.svcCtx.AppEnrollModel.IncrView(in.Aeid)
	// 更新总 投票和总浏览
	l.svcCtx.AppVotesActivityModel.IncrVotes(in.Actid)
	l.svcCtx.AppVotesActivityModel.IncrView(in.Actid)
	return &votes.VotesResp{
		Aeid: in.Aeid,
		Uid:  in.Uid,
		Auid: in.Auid,
		Beid: in.Beid,
	}, err
}

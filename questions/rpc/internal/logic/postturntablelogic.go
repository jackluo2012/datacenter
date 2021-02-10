package logic

import (
	"context"

	"datacenter/questions/model"
	"datacenter/questions/rpc/internal/svc"
	"datacenter/questions/rpc/questions"
	"datacenter/shared"

	"github.com/tal-tech/go-zero/core/logx"
	"github.com/tal-tech/go-zero/core/stores/redis"
)

type PostTurnTableLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPostTurnTableLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PostTurnTableLogic {
	return &PostTurnTableLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 抽奖
func (l *PostTurnTableLogic) PostTurnTable(in *questions.TurnTableReq) (*questions.AwardInfoResp, error) {

	//1. 加锁
	redisLock := redis.NewRedisLock(l.svcCtx.RedisConn, shared.GetUidAuidLockKey(in.Uid, in.Auid))
	// 2. 可选操作，设置 redislock 过期时间
	redisLock.SetExpire(shared.UidAuidLockExpire)
	if ok, err := redisLock.Acquire(); !ok || err != nil {
		return nil, shared.ErrorUserOperation
	}
	// 3. 释放锁
	defer redisLock.Release()

	//1. 检查用户是否已经抽过奖了

	//2. 获取抽奖的随机编码
	prizeCode := shared.RandInt(10000)

	//3.匹配是否中奖了
	awards, err := l.prize(in.ActivityId, int64(prizeCode))
	if err != nil {
		return nil, err
	}
	//如果是虚拟高品
	if awards.IsLottery == 0 {
		goto doDefault
	}
	err = l.svcCtx.QuestionsLotteriesModel.TurnTable(model.AppQuestionsLotteries{
		ActivityId: in.ActivityId,
		Uid:        in.Uid,
		Auid:       in.Auid,
		Ptyid:      in.Ptyid,
		Beid:       in.Beid,
		AwardId:    awards.Id,
		IsWinning:  awards.IsLottery,
	})
	//如果因为库存不足，各种问题报错，就随便给一个
	if err != nil {
		goto doDefault
	}

doDefault:
	{
		//随机写一个奖品
		awards, err = l.defaultPrize(in.ActivityId)
		l.svcCtx.QuestionsLotteriesModel.Insert(model.AppQuestionsLotteries{
			ActivityId: in.ActivityId,
			Uid:        in.Uid,
			Auid:       in.Auid,
			Ptyid:      in.Ptyid,
			Beid:       in.Beid,
			AwardId:    awards.Id,
			IsWinning:  awards.IsLottery,
		})
	}

	return &questions.AwardInfoResp{
		Id:         awards.Id,
		Beid:       awards.Beid,
		Ptyid:      awards.Ptyid,
		ActivityId: awards.ActivityId,
	}, nil
}

//检查 奖品信息
func (l *PostTurnTableLogic) prize(actid, prizeCode int64) (*model.AppQuestionsAwards, error) {
	var prizeGift *model.AppQuestionsAwards

	giftList, err := l.svcCtx.QuestionsAwardsModel.Find(actid)
	if err != nil {
		return nil, err
	}

	for _, gift := range giftList {
		if gift.StartProbability <= prizeCode && gift.EndProbability >= prizeCode {
			prizeGift = &gift
			break
		}
	}
	return prizeGift, nil
}

/**
 *	随机拿一个奖品吧
 */
func (l *PostTurnTableLogic) defaultPrize(actid int64) (*model.AppQuestionsAwards, error) {

	giftList, err := l.svcCtx.QuestionsAwardsModel.Find(actid)
	if err != nil {
		return nil, err
	}
	noLottery := make([]model.AppQuestionsAwards, 0)
	for _, gift := range giftList {
		// 中奖编码区间满足条件，说明可以中奖
		if gift.IsLottery == 0 {
			noLottery = append(noLottery, gift)
		}
	}
	return &noLottery[shared.RandInt(len(noLottery))], nil

}

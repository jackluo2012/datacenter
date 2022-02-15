package logic

import (
	"context"
	"database/sql"

	"datacenter/questions/model"
	"datacenter/questions/rpc/internal/svc"
	"datacenter/questions/rpc/questions"

	"github.com/zeromicro/go-zero/core/logx"
)

type PostConvertLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPostConvertLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PostConvertLogic {
	return &PostConvertLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 填写中奖记录
func (l *PostConvertLogic) PostConvert(in *questions.ConvertReq) (*questions.ConvertResp, error) {
	result, err := l.svcCtx.QuestionsConvertsModel.Insert(model.AppQuestionsConverts{
		LotteryId: in.LotteryId,
		Username:  in.Username,
		Ptyid:     in.Ptyid,
		Beid:      in.Beid,
		Uid:       in.Uid,
		Auid:      in.Auid,
		Phone:     sql.NullString{String: in.Phone, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	return &questions.ConvertResp{
		Id:        id,
		LotteryId: in.LotteryId,
	}, err
}

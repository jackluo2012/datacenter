package logic

import (
	"context"

	"datacenter/datasvc/model"
	"datacenter/datasvc/rpc/datasvc"
	"datacenter/datasvc/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListCallLogsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListCallLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListCallLogsLogic {
	return &ListCallLogsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 调用日志：beid/api_id/app_id/status 多维可选过滤 + 分页（时间倒序，分页参数在 model 层兜底）
func (l *ListCallLogsLogic) ListCallLogs(in *datasvc.CallLogListReq) (*datasvc.CallLogListReply, error) {
	list, total, err := l.svcCtx.CallLogModel.FindList(in.Beid, in.ApiId, in.AppId, in.Status, in.Page, in.Size)
	if err != nil {
		return nil, err
	}
	pbList := make([]*datasvc.ApiCallLog, 0, len(list))
	for _, item := range list {
		pbList = append(pbList, callLogToPb(item))
	}
	return &datasvc.CallLogListReply{List: pbList, Total: total}, nil
}

// callLogToPb model.ApiCallLog → pb ApiCallLog（created_at 输出毫秒时间戳）
func callLogToPb(item *model.ApiCallLog) *datasvc.ApiCallLog {
	if item == nil {
		return nil
	}
	return &datasvc.ApiCallLog{
		Id:        item.Id,
		ApiId:     item.ApiId,
		AppId:     item.AppId,
		Path:      item.Path,
		Status:    item.Status,
		ErrorCode: item.ErrorCode,
		RowsCnt:   item.RowsCnt,
		CostMs:    int32(item.CostMs),
		CacheHit:  item.CacheHit == 1,
		Ip:        item.Ip,
		CreatedAt: item.CreatedAt.UnixMilli(),
	}
}

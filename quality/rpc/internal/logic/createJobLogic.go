package logic

import (
	"context"
	"fmt"
	"strings"

	"datacenter/quality/model"
	"datacenter/quality/rpc/internal/svc"
	"datacenter/quality/rpc/quality"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateJobLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateJobLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateJobLogic {
	return &CreateJobLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建任务：校验 rule_ids 非空且每条规则存在（同 beid 租户内）
func (l *CreateJobLogic) CreateJob(in *quality.CreateJobReq) (*quality.JobReply, error) {
	if in.Beid <= 0 {
		return nil, model.ErrBeidRequired
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, ErrJobNameRequired
	}
	if in.DatasourceId <= 0 {
		return nil, ErrDatasourceRequired
	}
	ruleIds, err := ParseRuleIds(in.RuleIds)
	if err != nil {
		return nil, err
	}
	// 规则必须存在且属于同租户（FindListByIds 内部按 beid 过滤）
	if err := l.checkRulesExist(in.Beid, ruleIds); err != nil {
		return nil, err
	}
	id, err := l.svcCtx.JobModel.Insert(l.ctx, &model.Job{
		Name:         in.Name,
		DatasourceId: in.DatasourceId,
		DbName:       in.DbName,
		RuleIds:      in.RuleIds,
		Cron:         in.Cron,
		Beid:         in.Beid,
	})
	if err != nil {
		return nil, err
	}
	job, err := l.svcCtx.JobModel.FindOne(l.ctx, id, in.Beid)
	if err != nil {
		return nil, err
	}
	return &quality.JobReply{Job: jobToPb(job)}, nil
}

// checkRulesExist 校验规则 id 全部存在，缺失时返回缺失明细
func (l *CreateJobLogic) checkRulesExist(beid int64, ruleIds []int64) error {
	rules, err := l.svcCtx.RuleModel.FindListByIds(l.ctx, beid, ruleIds)
	if err != nil {
		return err
	}
	found := make(map[int64]struct{}, len(rules))
	for _, r := range rules {
		found[r.Id] = struct{}{}
	}
	var missing []int64
	for _, id := range ruleIds {
		if _, ok := found[id]; !ok {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("%w: 缺失id %v", ErrRuleNotFound, missing)
	}
	return nil
}

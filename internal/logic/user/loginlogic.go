package logic

import (
	"context"
	"time"

	"datacenter/internal/logic"
	"datacenter/internal/middleware"
	"datacenter/internal/svc"
	"datacenter/internal/types"
	"datacenter/ops/rpc/opsClient"
	"datacenter/user/rpc/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

// loginLogTimeout 登录日志异步落库的 RPC 超时保护
const loginLogTimeout = 3 * time.Second

type LoginLogic struct {
	logx.Logger
	ctx             context.Context
	svcCtx          *svc.ServiceContext
	datacenterLogic logic.DatacenterLogic
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) LoginLogic {
	return LoginLogic{
		Logger:          logx.WithContext(ctx),
		ctx:             ctx,
		svcCtx:          svcCtx,
		datacenterLogic: logic.NewDatacenterLogic(ctx, svcCtx),
	}
}

func (l *LoginLogic) Login(req types.LoginReq, beid, ptyid int64) (*types.UserReply, error) {

	reply, err := l.svcCtx.UserRpc.Login(l.ctx, &userclient.LoginReq{
		Mobile:   req.Mobile,
		Type:     req.Type,
		Password: req.Password,
	})
	if err != nil {
		// 登录失败同样落登录日志（等保要求：成功/失败均留痕）
		l.writeLoginLog(req.Mobile, 0, beid, false, err.Error())
		return nil, err
	}
	jwttoken, err := l.datacenterLogic.GetJwtToken(types.AppUser{Uid: reply.Uid, Ptyid: ptyid, Beid: beid})
	if err != nil {
		// 签发 token 失败也视为登录失败留痕
		l.writeLoginLog(req.Mobile, reply.Uid, beid, false, err.Error())
		return nil, err
	}
	// 登录成功落登录日志
	l.writeLoginLog(req.Mobile, reply.Uid, beid, true, "登录成功")
	return &types.UserReply{
		Auid:     reply.Auid,
		Beid:     beid,
		Ptyid:    ptyid,
		Uid:      reply.Uid,
		Username: reply.Username,
		Mobile:   reply.Mobile,
		Avator:   reply.Avator,
		JwtToken: jwttoken,
	}, nil
}

// writeLoginLog 异步写登录日志到 ops 服务（审计旁路）：失败仅记日志，不影响登录主流程。
// 客户端 ip/ua 由全局审计中间件注入 ctx（ClientIPCtxKey/UserAgentCtxKey）。
func (l *LoginLogic) writeLoginLog(username string, uid, beid int64, success bool, msg string) {
	ip, _ := l.ctx.Value(middleware.ClientIPCtxKey).(string)
	userAgent, _ := l.ctx.Value(middleware.UserAgentCtxKey).(string)
	go func() {
		defer func() {
			// 登录日志属旁路逻辑，任何异常不得影响登录
			if r := recover(); r != nil {
				logx.Errorf("[登录日志] 写登录日志 panic: %v, username=%s", r, username)
			}
		}()
		ctx, cancel := context.WithTimeout(context.Background(), loginLogTimeout)
		defer cancel()
		if _, err := l.svcCtx.OpsRpc.WriteLoginLog(ctx, &opsClient.WriteLoginLogReq{
			Username:  username,
			Uid:       uid,
			Ip:        ip,
			UserAgent: userAgent,
			Success:   success,
			Msg:       msg,
			Beid:      beid,
		}); err != nil {
			logx.Errorf("[登录日志] 写登录日志失败 username=%s success=%v: %v", username, success, err)
		}
	}()
}

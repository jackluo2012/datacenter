package logic

import (
	"context"
	"fmt"
	"strings"

	"datacenter/datasvc/model"
	"datacenter/datasvc/rpc/datasvc"
	"datacenter/datasvc/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// defaultAppQps 应用未显式配置限流时的默认 QPS（对齐 dc_api_app 表默认值）
const defaultAppQps int64 = 10

type RegisterAppLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterAppLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterAppLogic {
	return &RegisterAppLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 注册应用：生成 app_key（app_+随机 hex16）、app_secret（随机 hex32），
// secret 经 AES-GCM 加密入库，明文仅本次响应返回一次
func (l *RegisterAppLogic) RegisterApp(in *datasvc.RegisterAppReq) (*datasvc.AppReply, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: name 不能为空", ErrParamInvalid)
	}
	if in.RateLimitQps < 0 {
		return nil, fmt.Errorf("%w: rate_limit_qps 不能为负数", ErrParamInvalid)
	}
	qps := in.RateLimitQps
	if qps == 0 {
		qps = defaultAppQps //未配置时取默认限流
	}
	appKey, err := randomHex(8) //app_ + 16 位 hex
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSecretCrypto, err)
	}
	appKey = "app_" + appKey
	secret, err := randomHex(16) //32 位 hex
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSecretCrypto, err)
	}
	//AES-GCM 加密入库，明文不落库
	enc, err := encryptAppSecret(l.svcCtx, secret)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSecretCrypto, err)
	}
	data := &model.ApiApp{
		AppKey:       appKey,
		AppSecretEnc: enc,
		Name:         name,
		RateLimitQps: qps,
		IpWhitelist:  normalizeWhitelist(in.IpWhitelist),
		Beid:         in.Beid,
	}
	id, err := l.svcCtx.AppModel.Insert(data)
	if err != nil {
		return nil, err
	}
	data.Id = id
	l.Infof("注册应用成功 id=%d beid=%d app_key=%s", id, data.Beid, data.AppKey)
	return &datasvc.AppReply{App: appToPb(data), AppSecret: secret}, nil
}

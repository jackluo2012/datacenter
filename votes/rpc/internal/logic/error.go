package logic

import "errors"

var (
	ActivityDoesNotExist                 = errors.New("活动不存在")
	ActivityEnd                          = errors.New("活动已结束")
	ActivityDoesNotOpen                  = errors.New("活动未开启")
	RegistrationActivityDoesNotEnroll    = errors.New("活动还未开始报名")
	ActivityDoesNotStart                 = errors.New("活动还未开始")
	RegistrationActivityDoesNotEnrollEND = errors.New("报名活动未开启")
	EnrollFalt                           = errors.New("报名失败")
	YouHaveSignedUp                      = errors.New("你已经报名")
	ActivityDoesNotEnroll                = errors.New("活动还未开始投票")
	VotesLock                            = errors.New("当前投票处理中，请稍后重试")
	VotesFailt                           = errors.New("投票失败")

	EnrollNotExist = errors.New("投票的作品不存在")
)

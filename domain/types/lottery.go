package types

const (
	//抽奖流程状态
	LotteryStatusWait     = 0 //等待抽奖
	LotteryStatusGetPrize = 1 //获取奖品
	LotteryStatusDeduct   = 2 //扣除金钱
	LotteryStatusAward    = 3 //发奖
)

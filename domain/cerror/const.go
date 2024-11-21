package cerror

// system
var (
	ErrSystem     = NewError(10001, "系统错误，稍候再试")
	ErrBusy       = NewError(10002, "系统繁忙，稍候再试")
	ErrFrequently = NewError(10003, "请求过于频繁，请稍候再试")
	ErrParam      = NewError(10004, "参数错误")
	ErrTimeout    = NewError(10005, "请求超时")
	ErrDuplicate  = NewError(10006, "重复请求，请刷新后重试")
)

// account
var (
	ErrNotAccount = NewError(11001, "账号或密码错误")
	ErrPassword   = NewError(11002, "账号或密码错误")
	ErrEmailExist = NewError(11003, "邮箱已存在")
	ErrLogout     = NewError(11004, "未登录，请先登录")
)

// lottery
var (
	ErrLotteryConfig  = NewError(12001, "抽奖配置错误，请检查")
	ErrLotteryNoPrize = NewError(12002, "抽奖错误，没有奖品")
	ErrLotteryNoAct   = NewError(12003, "抽奖活动不存在，请刷新")
)

// asset
var (
	ErrAssetLess = NewError(13001, "资产不足")
	ErrItemLess  = NewError(13002, "物品不足")
)

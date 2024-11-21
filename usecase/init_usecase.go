package usecase

import (
	"github.com/linchengzhi/lottery/Infra/gpool"
	"github.com/linchengzhi/lottery/repository/mysql_repo"
	"github.com/linchengzhi/lottery/repository/redis_repo"
	"github.com/linchengzhi/lottery/usecase/asset_uc"
	"github.com/linchengzhi/lottery/usecase/lottery_uc"
	"go.uber.org/zap"
)

type UcAll struct {
	asset_uc.AssetUc
	lottery_uc.LotteryUc
}

func NewUcAll(log *zap.Logger, g *gpool.Pool, repoMysql mysql_repo.RepoMysql, repoRedis redis_repo.RepoRedis, repoStream redis_repo.RepoStream) UcAll {
	uc := new(UcAll)
	uc.AssetUc = asset_uc.NewAssetUc(log, repoMysql, repoRedis)
	uc.LotteryUc = lottery_uc.NewLotteryUc(log, g, repoMysql, repoRedis, repoStream, uc.AssetUc)
	return *uc
}

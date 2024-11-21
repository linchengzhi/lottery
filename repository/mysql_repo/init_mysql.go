package mysql_repo

import (
	"gorm.io/gorm"
)

type RepoMysql struct {
	UserAssetRepo
	UserAssetRecordRepo
	UserItemRepo

	LotteryDrawRecordRepo
	LotteryPrizeRecordRepo
}

func NewRepoMysql(db *gorm.DB) RepoMysql {
	repo := new(RepoMysql)

	repo.UserAssetRepo = NewUserAssetRepo(db)
	repo.UserAssetRecordRepo = NewUserAssetRecordRepo(db)
	repo.UserItemRepo = NewUserItemRepo(db)
	repo.LotteryDrawRecordRepo = NewLotteryDrawRecordRepo(db)
	repo.LotteryPrizeRecordRepo = NewLotteryPrizeRecordRepo(db)
	return *repo
}

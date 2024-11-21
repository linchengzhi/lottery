package redis_repo

import (
	"github.com/go-redis/redis/v8"
	"github.com/linchengzhi/lottery/Infra/database/redis_db"
	"github.com/linchengzhi/lottery/Infra/gpool"
	"github.com/linchengzhi/lottery/domain/dto"
	"github.com/linchengzhi/lottery/domain/types"
)

type RepoRedis struct {
	UserAssetCache
	UserItemCache
	LotteryRecordCache
}

type RepoStream struct {
	DrawRs  redis_db.IStream
	AwardRs redis_db.IStream
}

func NewRepoRedis(rd *redis.Client) RepoRedis {
	repo := new(RepoRedis)
	repo.UserAssetCache = NewUserAssetCache(rd)
	repo.UserItemCache = NewUserItemCache(rd)
	repo.LotteryRecordCache = NewLotteryRecordCache(rd)
	return *repo
}

// NewRepoStream 创建redis stream
func NewRepoStream(rd *redis.Client, pool *gpool.Pool, stream []dto.RedisStream) (RepoStream, error) {
	repo := new(RepoStream)
	for _, v := range stream {
		rs, err := redis_db.NewRedisStream(rd, pool, v)
		if err != nil {
			return *repo, err
		}
		switch v.Name {
		case types.StreamLottery:
			repo.DrawRs = rs
		case types.StreamAward:
			repo.AwardRs = rs
		}
	}
	return *repo, nil
}

package redis_repo

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"strconv"
	"time"
)

// 使用hash保存奖池
type IPrizePoolRd interface {
	Get(ctx context.Context, activityId int64) (map[int64]int64, error)
	DecrBy(ctx context.Context, activityId int64, data map[int64]int64) error
	Set(ctx context.Context, activityId int64, data map[int64]int64, ttl time.Duration) error
}

var keyPrizePoolHash = "lottery:" + "prize_pool:%d"

type PrizePoolRd struct {
	rd *redis.Client
}

func NewPrizePoolRd(rd *redis.Client) *PrizePoolRd {
	return &PrizePoolRd{rd: rd}
}

func (r *PrizePoolRd) Get(ctx context.Context, activityId int64) (map[int64]int64, error) {
	key := fmt.Sprintf(keyPrizePoolHash, activityId)
	data, err := r.rd.HGetAll(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}
	result := make(map[int64]int64)
	for k, v := range data {
		id, _ := strconv.ParseInt(k, 10, 64)
		num, _ := strconv.ParseInt(v, 10, 64)
		result[id] = num
	}
	return result, nil
}

func (r *PrizePoolRd) DecrBy(ctx context.Context, activityId int64, data map[int64]int64) error {
	key := fmt.Sprintf(keyPrizePoolHash, activityId)
	pipe := r.rd.Pipeline()
	for k, v := range data {
		field := fmt.Sprintf("%d", k)
		pipe.HIncrBy(ctx, key, field, -v)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *PrizePoolRd) Set(ctx context.Context, activityId int64, data map[int64]int64, ttl time.Duration) error {
	key := fmt.Sprintf(keyPrizePoolHash, activityId)
	// 检查键是否存在
	exists, err := r.rd.Exists(ctx, key).Result()
	if err != nil {
		return err
	}
	if exists > 0 {
		return nil
	}

	// 构造要写入的数据
	pipe := r.rd.Pipeline()
	for k, v := range data {
		field := fmt.Sprintf("%d", k)
		pipe.HSet(ctx, key, field, v)
	}
	_, err = pipe.Exec(ctx)
	if err != nil {
		return err
	}

	_, err = r.rd.Expire(ctx, key, ttl).Result()
	if err != nil {
		return err
	}
	return nil
}

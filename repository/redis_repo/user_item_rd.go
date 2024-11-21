package redis_repo

import (
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/go-redis/redis/v8"
	"time"
)

type UserItemCache struct {
	redisClient *redis.Client
	expiration  time.Duration
}

var keyUserItem = "asset:" + "user_item:%d"

func NewUserItemCache(rd *redis.Client) UserItemCache {
	return UserItemCache{
		redisClient: rd,
		expiration:  time.Duration(10) * time.Minute, // 10分钟过期
	}
}

func (uic *UserItemCache) Get(ctx context.Context, userId int64) (map[int64]int64, error) {
	key := fmt.Sprintf(keyUserItem, userId)
	data, err := uic.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // 缓存未命中
	} else if err != nil {
		return nil, err
	}

	var items map[int64]int64
	if err = sonic.Unmarshal([]byte(data), &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (uic *UserItemCache) Set(ctx context.Context, userId int64, items map[int64]int64) error {
	key := fmt.Sprintf(keyUserItem, userId)
	data, err := sonic.Marshal(items)
	if err != nil {
		return err
	}

	return uic.redisClient.Set(ctx, key, data, uic.expiration).Err()
}

func (uic *UserItemCache) Delete(ctx context.Context, userId int64) error {
	key := fmt.Sprintf(keyUserItem, userId)
	return uic.redisClient.Del(ctx, key).Err()
}

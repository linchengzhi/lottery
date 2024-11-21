package redis_repo

import (
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/go-redis/redis/v8"
	"github.com/linchengzhi/lottery/domain/entity"
	"time"
)

type UserAssetCache struct {
	redisClient *redis.Client
	expiration  time.Duration
}

var keyUserAsset = "asset:" + "user_asset:%d"

func NewUserAssetCache(rd *redis.Client) UserAssetCache {
	return UserAssetCache{
		redisClient: rd,
		expiration:  time.Duration(10) * time.Minute, // 10分钟过期
	}
}

func (uac *UserAssetCache) Get(ctx context.Context, userId int64) (*entity.UserAsset, error) {
	key := fmt.Sprintf(keyUserAsset, userId)
	data, err := uac.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // 缓存未命中
	} else if err != nil {
		return nil, err
	}

	var assets = new(entity.UserAsset)
	if err = sonic.Unmarshal([]byte(data), &assets); err != nil {
		return nil, err
	}
	return assets, nil
}

func (uac *UserAssetCache) Set(ctx context.Context, userId int64, asset *entity.UserAsset) error {
	key := fmt.Sprintf(keyUserAsset, userId)
	data, err := sonic.Marshal(asset)
	if err != nil {
		return err
	}
	return uac.redisClient.Set(ctx, key, data, uac.expiration).Err()
}

func (uac *UserAssetCache) Delete(ctx context.Context, userId int64) error {
	key := fmt.Sprintf(keyUserAsset, userId)
	return uac.redisClient.Del(ctx, key).Err()
}

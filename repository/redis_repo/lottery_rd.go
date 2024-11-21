package redis_repo

import (
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/go-redis/redis/v8"
	"github.com/linchengzhi/lottery/domain/dto"
	"time"
)

type ILotteryRecordRd interface {
	Get(ctx context.Context, requestId string) (*dto.DrawReq, error)
	Set(ctx context.Context, req *dto.DrawReq) error
	Del(ctx context.Context, requestId string) error
	//定时
	GetTimeout(ctx context.Context, callback func(req *dto.DrawReq) error) error
}

type LotteryRecordCache struct {
	rdb        *redis.Client
	expiration time.Duration
	waitTime   time.Duration // 等待时间
	timeout    time.Duration
}

// 2. Redis key 设计
const (
	keyLotteryParam = "lottery:param:%s" //抽奖参数 -- 唯一键

	keyLotteryRecord = "lottery:draw:record" // 抽奖记录 唯一键-时间戳 用于获取超时数据
)

func NewLotteryRecordCache(rdb *redis.Client) LotteryRecordCache {
	return LotteryRecordCache{
		rdb:        rdb,
		expiration: time.Duration(24) * time.Hour,   // 1小时过期
		waitTime:   time.Duration(5) * time.Second,  // 5秒等待
		timeout:    time.Duration(60) * time.Second, // 60秒超时
	}
}

func (r *LotteryRecordCache) Get(ctx context.Context, requestId string) (*dto.DrawReq, error) {
	key := fmt.Sprintf(keyLotteryParam, requestId)
	data, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	var result = new(dto.DrawReq)
	err = sonic.Unmarshal([]byte(data), result)
	return result, nil
}

func (r *LotteryRecordCache) Set(ctx context.Context, req *dto.DrawReq) error {
	key := fmt.Sprintf(keyLotteryParam, req.RequestId)
	data, _ := sonic.Marshal(req)
	pip := r.rdb.Pipeline()

	pip.Set(ctx, key, data, r.expiration)

	pip.ZAdd(ctx, keyLotteryRecord, &redis.Z{float64(req.RequestTime.Unix()), req.RequestId})

	_, err := pip.Exec(ctx)
	return err
}

// 定时获取超时30秒的抽奖记录，如果没有等待五秒，如果有，则回调函数，成功后删除数据，然后继续获取下一个
// GetTimeout 定时获取超时记录并处理
func (r *LotteryRecordCache) GetTimeout(ctx context.Context, callback func(req *dto.DrawReq) error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// 获取30秒前的记录
			now := time.Now().Add(-r.timeout)
			records, err := r.rdb.ZRangeByScore(ctx, keyLotteryRecord, &redis.ZRangeBy{
				Min:    "0",
				Max:    fmt.Sprintf("%d", now.Unix()),
				Offset: 0,
				Count:  1,
			}).Result()

			if err != nil {
				return fmt.Errorf("get timeout records error: %w", err)
			}

			// 没有超时记录，等待5秒
			if len(records) == 0 {
				time.Sleep(r.waitTime)
				continue
			}

			// 获取详细数据
			requestId := records[0]
			req, err := r.Get(ctx, requestId)
			if err != nil {
				return fmt.Errorf("get record detail error: %w", err)
			}

			// 可能已经被删除
			if req == nil {
				// 清理zset中的数据
				r.rdb.ZRem(ctx, keyLotteryRecord, requestId)
				continue
			}

			// 执行回调
			err = callback(req)
			if err != nil {
				time.Sleep(1 * time.Second)
				return fmt.Errorf("callback error: %w", err)
			}
		}
	}
}

// Del 删除抽奖记录和参数
func (r *LotteryRecordCache) Del(ctx context.Context, requestId string) error {
	pip := r.rdb.Pipeline()

	// 删除抽奖参数
	key := fmt.Sprintf(keyLotteryParam, requestId)
	pip.Del(ctx, key)

	// 删除有序集合中的记录
	pip.ZRem(ctx, keyLotteryRecord, requestId)

	_, err := pip.Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete lottery record error: %w", err)
	}
	return nil
}

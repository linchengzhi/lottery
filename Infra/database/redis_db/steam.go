package redis_db

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/linchengzhi/lottery/Infra/gpool"
	"github.com/linchengzhi/lottery/domain/dto"
	"log"
	"time"
)

// SteamCallBack 是一个全局的回调函数类型，用于处理读取到的消息

// IStream 定义了 Redis Stream 的接口
type IStream interface {
	Add(data string) (string, error)
	Get(callback func(message redis.XMessage) error)
	Ack(messageID string) error
	GetPending() ([]redis.XMessage, error)
	RetryTimeoutMessages(handler func(message redis.XMessage) error)
}

// RedisSteam 是 Redis Stream 的具体实现
type RedisStream struct {
	rdb          *redis.Client
	name         string // 消息队列名称
	group        string // 消费者组名称
	consumerName string // 消费者名称
	pool         *gpool.Pool
	timeout      time.Duration // 消息读取超时时间
}

// NewRedisSteam 创建一个新的 RedisStream 实例，并确保消费者组存在
func NewRedisStream(rdb *redis.Client, pool *gpool.Pool, stream dto.RedisStream) (IStream, error) {
	rs := &RedisStream{
		rdb:          rdb,
		name:         stream.Name,
		group:        stream.Group,
		consumerName: stream.Consumer,
		pool:         pool,
		timeout:      time.Second * 30, // 30秒
	}
	err := rs.create()
	if err != nil {
		return nil, err
	}
	return rs, nil
}

// create 确保 Redis Stream 和消费者组存在
func (rs *RedisStream) create() error {
	_, err := rs.rdb.XGroupCreateMkStream(context.TODO(), rs.name, rs.group, "$").Result()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("failed to create consumer group: %v", err)
	}
	return nil
}

// Add 将用户ID和数据写入Redis Stream
func (rs *RedisStream) Add(data string) (string, error) {
	// 使用 XAdd 将消息推入 Redis Stream
	values := make(map[string]interface{})
	values["data"] = data
	result, err := rs.rdb.XAdd(context.TODO(), &redis.XAddArgs{
		Stream: rs.name,
		Values: values,
	}).Result()

	if err != nil {
		return "", err
	}
	return result, nil
}

// Get 从 Redis Stream 中读取消息，并通过回调函数返回结果
func (rs *RedisStream) Get(callback func(message redis.XMessage) error) {
	for {
		// 从 Redis Stream 中读取消息
		streams, err := rs.rdb.XReadGroup(context.TODO(), &redis.XReadGroupArgs{
			Group:    rs.group,
			Consumer: rs.consumerName,
			Streams:  []string{rs.name, ">"},
			Count:    1,
			Block:    time.Second * 1, // 阻塞1秒，等待消息
		}).Result()

		if err != nil {
			if err != redis.Nil {
				log.Printf("Error reading from stream: %v\n", err)
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// 遍历读取到的消息
		for _, stream := range streams {
			for _, message := range stream.Messages {
				// 将消息通过回调函数返回
				rs.pool.Submit(func() {
					err = callback(message) // 通过协程池执行回调
					if err == nil {
						rs.Ack(message.ID)
					}
				})
			}
		}
	}
}

// Ack 确认消息已被处理
func (rs *RedisStream) Ack(messageID string) error {
	_, err := rs.rdb.XAck(context.TODO(), rs.name, rs.group, messageID).Result()
	if err != nil {
		return fmt.Errorf("failed to ack message: %v", err)
	}
	return nil
}

func (rs *RedisStream) GetPending() ([]redis.XMessage, error) {
	// 使用pipeline减少网络往返
	pipe := rs.rdb.Pipeline()

	// 获取待处理消息
	pendingCmd := pipe.XPendingExt(context.TODO(), &redis.XPendingExtArgs{
		Stream: rs.name,
		Group:  rs.group,
		Start:  "-",
		End:    "+",
		Count:  100,
	})

	_, err := pipe.Exec(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to get pending messages: %v", err)
	}

	pending, err := pendingCmd.Result()
	if err != nil {
		return nil, err
	}

	// 收集需要认领的消息ID
	var claimIDs []string
	for _, p := range pending {
		if p.Idle.Milliseconds() >= rs.timeout.Milliseconds() {
			claimIDs = append(claimIDs, p.ID)
		}
	}

	// 如果没有需要认领的消息，直接返回
	if len(claimIDs) == 0 {
		return nil, nil
	}

	// 批量认领消息
	messages, err := rs.rdb.XClaim(context.TODO(), &redis.XClaimArgs{
		Stream:   rs.name,
		Group:    rs.group,
		Consumer: rs.consumerName,
		MinIdle:  rs.timeout,
		Messages: claimIDs,
	}).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to claim messages: %v", err)
	}

	return messages, nil
}

func (rs *RedisStream) RetryTimeoutMessages(handler func(message redis.XMessage) error) {
	ticker := time.NewTicker(rs.timeout / 5)
	defer ticker.Stop()

	for range ticker.C {
		messages, err := rs.GetPending()
		if err != nil {
			continue
		}

		for _, msg := range messages {
			rs.pool.Submit(func() {
				err = handler(msg)
				if err != nil {
					return
				}

				rs.Ack(msg.ID)
			})
		}
	}
}

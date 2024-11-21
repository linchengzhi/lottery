package middleware

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/linchengzhi/lottery/domain/cerror"
	"github.com/linchengzhi/lottery/util"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/ratelimit"
)

// 重复请求限制中间件 1秒1次
func RepeatedLimitMiddleware(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户标识(可以是用户ID、IP等)
		userKey := c.ClientIP() // 这里使用IP作为示例

		// Redis key
		key := fmt.Sprintf("system:ratelimit:%s", userKey)

		ctx := context.Background()

		// 使用 Redis MULTI/EXEC 事务保证原子性
		pipe := rdb.Pipeline()

		// 获取当前计数
		countCmd := pipe.Get(ctx, key)
		// 增加计数
		pipe.Incr(ctx, key)
		// 设置过期时间(如果key不存在)
		pipe.Expire(ctx, key, time.Second)

		// 执行事务
		_, err := pipe.Exec(ctx)
		if err != nil && err != redis.Nil {
			util.RespondErr(c, cerror.ErrBusy)
			c.Abort()
			return
		}

		// 获取当前请求数
		count, _ := countCmd.Int64()

		// 判断是否超过限制
		if count >= 1 { // 已经有一次请求了
			util.RespondErr(c, cerror.ErrFrequently)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitMiddleware 是一个限流中间件
func RateLimitMiddleware(rps int, maxQueueSize int64) gin.HandlerFunc {
	// 创建一个限流器，每秒允许 rps 次请求
	limiter := ratelimit.New(rps)

	// 当前的等待队列大小
	var currentQueueSize int64

	return func(c *gin.Context) {
		// 检查当前队列是否超过最大允许值
		if atomic.LoadInt64(&currentQueueSize) >= maxQueueSize {
			util.RespondErr(c, cerror.ErrBusy) // 返回 "系统繁忙"
			c.Abort()                          // 终止请求
			return
		}

		// 增加排队请求数
		atomic.AddInt64(&currentQueueSize, 1)

		// 获取下一个令牌
		now := limiter.Take()

		// 处理请求
		c.Next()

		// 请求处理完毕，减少排队请求数
		atomic.AddInt64(&currentQueueSize, -1)

		// 计算处理请求的耗时
		_ = time.Since(now)
	}
}

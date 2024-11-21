package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/linchengzhi/lottery/domain/cerror"
	"github.com/linchengzhi/lottery/util"
	"time"
)

// Middleware: 使用Redis进行requestId去重
func RequestIdMiddleware(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取requestId，如果没有则生成一个新的
		requestId := c.GetHeader("request_id")
		if requestId == "" {
			requestId = uuid.New().String()
			c.Request.Header.Set("request_id", requestId)
		}

		key := "system:request_id:" + requestId

		// 检查requestId是否存在于Redis中
		ctx := context.Background()
		exists, err := rdb.Exists(ctx, key).Result()
		if err != nil {
			util.RespondErr(c, cerror.ErrBusy)
			c.Abort()
			return
		}

		if exists > 0 {
			// 如果存在，说明这是重复的请求，返回错误响应
			util.RespondErr(c, cerror.ErrFrequently)
			c.Abort()
			return
		}

		// 将requestId存入Redis，并设置过期时间
		err = rdb.Set(ctx, key, "1", 5*time.Minute).Err()
		if err != nil {
			util.RespondErr(c, cerror.ErrBusy)
			c.Abort()
			return
		}

		// 继续处理请求
		c.Next()
	}
}

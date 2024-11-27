package redis_db

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

func NewRedis(host, password string, db int) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         host,
		Password:     password, // no password set
		DB:           db,       // use default Db
		PoolSize:     200,      // 连接池大小
		MinIdleConns: 20,       // 最小空闲连接数
		MaxConnAge:   time.Minute * 3,
	})

	err := rdb.Ping(context.TODO()).Err()
	if err != nil {
		return nil, err
	}
	return rdb, nil
}

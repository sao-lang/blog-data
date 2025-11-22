package redis

import (
	"context"

	re "github.com/redis/go-redis/v11"
)

type Config struct {
	Addr     string
	Password string
	DB       int
}

type RedisClient struct {
	client *re.Client
}

var defaultClient *RedisClient

// 初始化客户端
func NewClient(cfg Config) *RedisClient {
	rdb := re.NewClient(&re.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	defaultClient = &RedisClient{client: rdb}
	return defaultClient
}

// 获取默认客户端
func Default() *RedisClient {
	return defaultClient
}

// Ping 测试连接
func (r *RedisClient) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

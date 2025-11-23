package redis

import (
	"context"

	re "github.com/redis/go-redis/v9"
)

type Config struct {
	Addr     string
	Password string
	DB       int
}

type Client struct {
	client *re.Client
}

var defaultClient *Client

// 初始化客户端
func NewClient(cfg Config) *Client {
	rdb := re.NewClient(&re.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	defaultClient = &Client{client: rdb}
	return defaultClient
}

// 获取默认客户端
func Default() *Client {
	return defaultClient
}

// Ping 测试连接
func (r *Client) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

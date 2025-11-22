// Set 设置键值
package redis

import (
	"context"
	"time"
)

func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// Get 获取键值
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Delete 删除键
func (r *RedisClient) Delete(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	res, err := r.client.Exists(ctx, key).Result()
	return res > 0, err
}

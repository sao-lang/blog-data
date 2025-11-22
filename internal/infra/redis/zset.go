// ZAdd 添加有序集合
package redis

import (
	"context"

	re "github.com/redis/go-redis/v11"
)

func (r *RedisClient) ZAdd(ctx context.Context, key string, members ...*re.Z) error {
	return r.client.ZAdd(ctx, key, members...).Err()
}

// ZRange 获取有序集合
func (r *RedisClient) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.ZRange(ctx, key, start, stop).Result()
}

// ZRem 删除成员
func (r *RedisClient) ZRem(ctx context.Context, key string, members ...interface{}) error {
	return r.client.ZRem(ctx, key, members...).Err()
}

// ZScore 获取成员分数
func (r *RedisClient) ZScore(ctx context.Context, key, member string) (float64, error) {
	return r.client.ZScore(ctx, key, member).Result()
}

// ZCard 获取集合元素数量
func (r *RedisClient) ZCard(ctx context.Context, key string) (int64, error) {
	return r.client.ZCard(ctx, key).Result()
}

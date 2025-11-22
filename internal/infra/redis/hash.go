// HSet 设置 hash
package redis

import "context"

func (r *RedisClient) HSet(ctx context.Context, key string, values map[string]interface{}) error {
	return r.client.HSet(ctx, key, values).Err()
}

// HGet 获取 hash 中字段
func (r *RedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	return r.client.HGet(ctx, key, field).Result()
}

// HGetAll 获取整个 hash
func (r *RedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.client.HGetAll(ctx, key).Result()
}

// HDel 删除字段
func (r *RedisClient) HDel(ctx context.Context, key string, fields ...string) error {
	return r.client.HDel(ctx, key, fields...).Err()
}

// HExists 判断字段是否存在
func (r *RedisClient) HExists(ctx context.Context, key, field string) (bool, error) {
	return r.client.HExists(ctx, key, field).Result()
}

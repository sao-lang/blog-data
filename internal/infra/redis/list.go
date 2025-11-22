package redis

import "context"

// LPush 插入列表头
func (r *RedisClient) LPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.LPush(ctx, key, values...).Err()
}

// RPush 插入列表尾
func (r *RedisClient) RPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.RPush(ctx, key, values...).Err()
}

// LPop 弹出列表头
func (r *RedisClient) LPop(ctx context.Context, key string) (string, error) {
	return r.client.LPop(ctx, key).Result()
}

// RPop 弹出列表尾
func (r *RedisClient) RPop(ctx context.Context, key string) (string, error) {
	return r.client.RPop(ctx, key).Result()
}

// LRange 获取列表
func (r *RedisClient) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return r.client.LRange(ctx, key, start, stop).Result()
}

// LLen 获取列表长度
func (r *RedisClient) LLen(ctx context.Context, key string) (int64, error) {
	return r.client.LLen(ctx, key).Result()
}

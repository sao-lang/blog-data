// Tx 批量执行操作
package redis

import (
	"context"

	re "github.com/redis/go-redis/v11"
)

func (r *RedisClient) Tx(ctx context.Context, fn func(pipe re.Pipeliner) error) error {
	_, err := r.client.TxPipelined(ctx, fn)
	return err
}

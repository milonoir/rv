package scanner

import (
	"context"

	"github.com/go-redis/redis/v8"
	r "github.com/milonoir/rv/redis"
)

// executor executes read-only Redis commands.
type executor struct {
	rc *redis.Client
}

// newExecutor returns a fully configured executor.
func newExecutor(rc *redis.Client) *executor {
	return &executor{
		rc: rc,
	}
}

// Execute implements the Executor interface.
func (e *executor) Execute(ctx context.Context, key string, rt r.DataType) (interface{}, error) {
	switch rt {
	case r.TypeList:
		return e.getList(ctx, key)
	case r.TypeSet:
		return e.getSet(ctx, key)
	case r.TypeSortedSet:
		return e.getSortedSet(ctx, key)
	case r.TypeHash:
		return e.getHash(ctx, key)
	default:
		// Assuming everything else is a single key.
		return e.getKey(ctx, key)
	}
}

func (e *executor) getKey(ctx context.Context, key string) ([]string, error) {
	v, err := e.rc.Get(ctx, key).Result()
	return []string{v}, err
}

func (e *executor) getList(ctx context.Context, key string) ([]string, error) {
	return e.rc.LRange(ctx, key, 0, -1).Result()
}

func (e *executor) getSet(ctx context.Context, key string) ([]string, error) {
	return e.rc.SMembers(ctx, key).Result()
}

func (e *executor) getSortedSet(ctx context.Context, key string) ([]redis.Z, error) {
	return e.rc.ZRangeWithScores(ctx, key, 0, -1).Result()
}

func (e *executor) getHash(ctx context.Context, key string) (map[string]string, error) {
	return e.rc.HGetAll(ctx, key).Result()
}

package category

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/umuttopalak/orders-api/model"
)

type RedisRepo struct {
	Client *redis.Client
}

type FindAllPage struct {
	Size   uint64
	Offset uint64
}

type FindResult struct {
	Categories []model.Category
	Cursor     uint64
}

var ErrNotExist = errors.New("category does not exist")

func CategoryIDKey(id uint64) string {
	return fmt.Sprintf("category:%d", id)
}

func (r *RedisRepo) Insert(ctx context.Context, category model.Category) error {
	data, err := json.Marshal(category)
	if err != nil {
		return fmt.Errorf("failed to encode category: %w", err)
	}

	key := CategoryIDKey(category.CategoryID)

	txn := r.Client.TxPipeline()

	res := txn.SetNX(ctx, key, string(data), 0)
	if err := res.Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to set: %w", err)
	}

	if err := txn.SAdd(ctx, "categories", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to add category to categories: %w", err)
	}

	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("failed to exec: %w", err)
	}

	return nil
}

func (r *RedisRepo) FindByID(ctx context.Context, id uint64) (model.Category, error) {
	key := CategoryIDKey(id)

	value, err := r.Client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return model.Category{}, ErrNotExist
	} else if err != nil {
		return model.Category{}, fmt.Errorf("get category: %w", err)
	}

	var category model.Category
	err = json.Unmarshal([]byte(value), &category)
	if err != nil {
		return model.Category{}, fmt.Errorf("failed to decode category json: %w", err)
	}

	return category, nil
}

func (r *RedisRepo) DeleteByID(ctx context.Context, id uint64) error {
	key := CategoryIDKey(id)

	txn := r.Client.TxPipeline()

	err := txn.Del(ctx, key).Err()
	if errors.Is(err, redis.Nil) {
		txn.Discard()
		return ErrNotExist
	} else if err != nil {
		txn.Discard()
		return fmt.Errorf("get category: %w", err)
	}

	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("failed to exec: %w", err)
	}

	return nil
}

func (r *RedisRepo) Update(ctx context.Context, category model.Category) error {
	data, err := json.Marshal(category)
	if err != nil {
		return fmt.Errorf("failed to encode category: %w", err)
	}

	key := CategoryIDKey(category.CategoryID)

	err = r.Client.SetXX(ctx, key, string(data), 0).Err()
	if errors.Is(err, redis.Nil) {
		return ErrNotExist
	} else if err != nil {
		return fmt.Errorf("set category: %w", err)
	}

	return nil
}

func (r *RedisRepo) FindAll(ctx context.Context, page FindAllPage) (FindResult, error) {
	res := r.Client.SScan(ctx, "categories", page.Offset, "*", int64(page.Size))

	keys, cursor, err := res.Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get category id's: %w", err)
	}

	if len(keys) == 0 {
		return FindResult{
			Categories: []model.Category{},
		}, nil
	}

	xs, err := r.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get categories: %w", err)
	}

	categories := make([]model.Category, len(xs))

	for i, x := range xs {
		x := x.(string)

		var category model.Category

		err := json.Unmarshal([]byte(x), &category)
		if err != nil {
			return FindResult{}, fmt.Errorf("failed to decode categories: %w", err)
		}

		categories[i] = category
	}

	return FindResult{
		Categories: categories,
		Cursor:     cursor,
	}, nil
}

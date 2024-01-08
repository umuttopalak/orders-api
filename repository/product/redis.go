package product

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

func ProductIDKey(id uint64) string {
	return fmt.Sprintf("product:%d", id)
}

type FindAllPage struct {
	Size   uint64
	Offset uint64
}

type FindResult struct {
	Products []model.Product
	Cursor   uint64
}

var ErrNotExist = errors.New("product does not exist")

func (r *RedisRepo) Insert(ctx context.Context, Product model.Product) error {
	data, err := json.Marshal(Product)
	if err != nil {
		return fmt.Errorf("failed to encode product: %w", err)
	}

	key := ProductIDKey(Product.ProductID)

	txn := r.Client.TxPipeline()

	res := r.Client.SetNX(ctx, key, string(data), 0)

	if err := res.Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to set: %w", err)
	}

	if err := txn.SAdd(ctx, "products", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to add Product to products: %w", err)
	}

	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("failed to exec: %w", err)
	}

	return nil
}

func (r *RedisRepo) FindByID(ctx context.Context, id uint64) (model.Product, error) {
	key := ProductIDKey(id)

	value, err := r.Client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return model.Product{}, ErrNotExist
	} else if err != nil {
		return model.Product{}, fmt.Errorf("get product: %w", err)
	}

	var Product model.Product
	err = json.Unmarshal([]byte(value), &Product)
	if err != nil {
		return model.Product{}, fmt.Errorf("failed to decode product json: %w", err)
	}

	return Product, nil
}

func (r *RedisRepo) DeleteByID(ctx context.Context, id uint64) error {
	key := ProductIDKey(id)

	txn := r.Client.TxPipeline()

	err := txn.Del(ctx, key).Err()
	if errors.Is(err, redis.Nil) {
		txn.Discard()
		return ErrNotExist
	} else if err != nil {
		txn.Discard()
		return fmt.Errorf("get product: %w", err)
	}

	if err := txn.SRem(ctx, "products", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to remove product from products: %w", err)
	}

	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("failed to exec: %w", err)
	}

	return nil
}

func (r *RedisRepo) Update(ctx context.Context, Product model.Product) error {
	data, err := json.Marshal(Product)
	if err != nil {
		return fmt.Errorf("Failed to encode product: %w", err)
	}

	key := ProductIDKey(Product.ProductID)

	err = r.Client.SetXX(ctx, key, string(data), 0).Err()
	if errors.Is(err, redis.Nil) {
		return ErrNotExist
	} else if err != nil {
		return fmt.Errorf("set product: %w", err)
	}

	return nil
}

func (r *RedisRepo) FindAll(ctx context.Context, page FindAllPage) (FindResult, error) {
	res := r.Client.SScan(ctx, "products", page.Offset, "*", int64(page.Size))

	keys, cursor, err := res.Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get Product id's: %w", err)
	}

	if len(keys) == 0 {
		return FindResult{
			Products: []model.Product{},
		}, nil
	}

	xs, err := r.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get products: %w", err)
	}

	products := make([]model.Product, len(xs))

	for i, x := range xs {
		x := x.(string)
		var Product model.Product

		err := json.Unmarshal([]byte(x), &Product)
		if err != nil {
			return FindResult{}, fmt.Errorf("failed to decode products json: %w", err)
		}

		products[i] = Product
	}

	return FindResult{
		Products: products,
		Cursor:   cursor,
	}, nil
}

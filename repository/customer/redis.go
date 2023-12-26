package customer

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
	Customers []model.Customer
	Cursor    uint64
}

var ErrNotExist = errors.New("customer does not exist")

func CustomerIDKey(id uint64) string {
	return fmt.Sprintf("customer:%d", id)
}

func (r *RedisRepo) Insert(ctx context.Context, customer model.Customer) error {
	data, err := json.Marshal(customer)
	if err != nil {
		return fmt.Errorf("failed to encode customer: %w", err)
	}

	key := CustomerIDKey(customer.CustomerID)

	txn := r.Client.TxPipeline()

	res := txn.SetNX(ctx, key, string(data), 0)
	if err := res.Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to set: %w", err)
	}

	if err := txn.SAdd(ctx, "customers", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to add customer to customers: %w", err)
	}

	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("failed to exec: %w", err)
	}

	return nil
}

func (r *RedisRepo) FindByID(ctx context.Context, id uint64) (model.Customer, error) {
	key := CustomerIDKey(id)

	value, err := r.Client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return model.Customer{}, ErrNotExist
	} else if err != nil {
		return model.Customer{}, fmt.Errorf("get customer: %w", err)
	}

	var customer model.Customer
	err = json.Unmarshal([]byte(value), &customer)
	if err != nil {
		return model.Customer{}, fmt.Errorf("failed to decode customer json: %w", err)
	}

	return customer, nil
}

func (r *RedisRepo) DeleteByID(ctx context.Context, id uint64) error {
	key := CustomerIDKey(id)

	txn := r.Client.TxPipeline()

	err := txn.Del(ctx, key).Err()
	if errors.Is(err, redis.Nil) {
		txn.Discard()
		return ErrNotExist
	} else if err != nil {
		txn.Discard()
		return fmt.Errorf("get customer: %w", err)
	}

	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("failed to exec: %w", err)
	}

	return nil
}

func (r *RedisRepo) Update(ctx context.Context, customer model.Customer) error {
	data, err := json.Marshal(customer)
	if err != nil {
		return fmt.Errorf("failed to encode customer: %w", err)
	}

	key := CustomerIDKey(customer.CustomerID)

	err = r.Client.SetXX(ctx, key, string(data), 0).Err()
	if errors.Is(err, redis.Nil) {
		return ErrNotExist
	} else if err != nil {
		return fmt.Errorf("set customer: %w", err)
	}

	return nil
}

func (r *RedisRepo) FindAll(ctx context.Context, page FindAllPage) (FindResult, error) {
	res := r.Client.SScan(ctx, "customers", page.Offset, "*", int64(page.Size))

	keys, cursor, err := res.Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get order id's: %w", err)
	}

	if len(keys) == 0 {
		return FindResult{
			Customers: []model.Customer{},
		}, nil
	}

	xs, err := r.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get customers: %w", err)
	}

	customers := make([]model.Customer, len(xs))

	for i, x := range xs {
		x := x.(string)
		var customer model.Customer

		err := json.Unmarshal([]byte(x), &customer)
		if err != nil {
			return FindResult{}, fmt.Errorf("failed to decode customer: %w", err)
		}

		customers[i] = customer
	}

	return FindResult{
		Customers: customers,
		Cursor:    cursor,
	}, nil
}

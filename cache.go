package aidns

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"
)

type Locker struct {
	client    *redis.Client
	redislock *redislock.Client
	ctx       context.Context
	retry     int
	retryTime time.Duration
	lockTime  time.Duration
	cacheTtl  time.Duration
}

func NewLocker(ctx context.Context, ad AiDNS, client *redis.Client) *Locker {
	return &Locker{
		client:    client,
		redislock: redislock.New(client),
		ctx:       ctx,
		retry:     10,
		retryTime: 10 * time.Millisecond,
		lockTime:  3 * time.Second,
		cacheTtl:  ad.RedisTTL,
	}
}

func (s Locker) GetCache(key string, v any, getFunc func() (any, error)) error {
	retry := 0
	for {
		err := s.Get(key, v, getFunc)
		if err != nil {
			time.Sleep(s.retryTime)
			retry++
			if retry >= s.retry {
				return errors.New("records is empty")
			}
		} else {
			return nil
		}
	}
}

func (s Locker) Get(key string, v any, getFunc func() (any, error)) error {
	get := s.client.Get(s.ctx, key)
	err := get.Err()
	if err != nil {
		// 缓存不存在
		var lock *redislock.Lock
		lock, err = s.redislock.Obtain(s.ctx, key+"-dblock", 10*time.Second, nil)
		switch {
		case errors.Is(err, redislock.ErrNotObtained):
			return errors.New("wait try")
		case err != nil:
			return errors.New("has lock")
		default:
			// 不要忘记推迟发布。
			defer lock.Release(s.ctx)
			// 查询数据
			var data any
			data, err = getFunc()
			if err != nil {
				return err
			}
			var marshal []byte
			marshal, err = json.Marshal(data)
			if err != nil {
				return err
			}
			err = s.client.Set(s.ctx, key, string(marshal), 10*time.Minute).Err()
			if err != nil {
				return err
			}
			err = json.Unmarshal(marshal, v)
			if err != nil {
				return err
			}
		}
		return nil
	}
	err = json.Unmarshal([]byte(get.Val()), v)
	if err != nil {
		return err
	}
	return nil
}

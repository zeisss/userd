package storage

import (
	"github.com/garyburd/redigo/redis"
	"github.com/juju/errgo"

	"time"
)

type redisKeyValueDriver struct {
	Pool  *redis.Pool
	Users *redisIndex
}

func NewRedisStorage(address string, maxIdle, maxActive int, timeout time.Duration, password string) *keyValueStorage {
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", address)
			if err != nil {
				return nil, err
			}

			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		MaxIdle:     maxIdle,
		MaxActive:   maxActive,
		IdleTimeout: timeout,
	}

	return newKeyValueStorage(&redisKeyValueDriver{
		Pool: pool,
		Users: &redisIndex{pool, func(key string) string {
			return "user:" + key
		}},
	})
}

func (r *redisKeyValueDriver) Set(userID, userJson string) error {
	return r.Users.Put(userID, userJson)
}

func (r *redisKeyValueDriver) Lookup(userID string) (string, bool, error) {
	return r.Users.Lookup(userID)
}

func (r *redisKeyValueDriver) Index(name string) keyValueIndex {
	return &redisIndex{Pool: r.Pool, Key: func(key string) string {
		return name + ":" + key
	}}
}

type redisIndex struct {
	Pool *redis.Pool

	Key func(key string) string
}

func (index *redisIndex) Put(key, value string) error {
	con := index.Pool.Get()
	defer con.Close()

	_, err := redis.String(con.Do("SET", index.Key(key), value))
	if err != nil {
		return errgo.Mask(err)
	}
	return nil
}

func (index *redisIndex) Remove(key string) error {
	con := index.Pool.Get()
	defer con.Close()

	_, err := redis.Int(con.Do("DEL", index.Key(key)))
	if err != nil {
		return errgo.Mask(err)
	}
	return nil
}

func (index *redisIndex) Lookup(key string) (string, bool, error) {
	con := index.Pool.Get()
	defer con.Close()

	value, err := redis.String(con.Do("GET", index.Key(key)))
	if err != nil {
		if err == redis.ErrNil {
			return "", false, nil
		}
		return "", false, errgo.Mask(err)
	}
	return value, true, nil
}

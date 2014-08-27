package main

import (
	"github.com/garyburd/redigo/redis"

	flag "github.com/ogier/pflag"

	"strings"
	"time"
)

func Dialer(address, password string) func() (redis.Conn, error) {
	// Be nice to docker and get rid of the protocol
	if strings.HasPrefix(address, "tcp://") {
		address = address[6:]
	}
	return func() (redis.Conn, error) {
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
	}
}

func NewPool(address, password string, maxIdle, maxActive int, timeout time.Duration) *redis.Pool {
	Dial := Dialer(address, password)

	return &redis.Pool{
		MaxIdle:     maxIdle,
		MaxActive:   maxActive,
		Dial:        Dial,
		IdleTimeout: timeout,
	}
}

var (
	redisAddress     = flag.String("redis-address", "localhost:6370", "Address to connect to.")
	redisPassword    = flag.String("redis-auth", "", "Password to send when establishing a connection.")
	redisMaxIdle     = flag.Int("redis-max-idle", 20, "Maximum number of idle connections before closing.")
	redisMaxActive   = flag.Int("redis-max-active", 20, "Maximum number of open connections.")
	redisIdleTimeout = flag.Int("redis-idle", 10*60, "Seconds connections can be idle before closing.")

	pool *redis.Pool
)

func RedisPool() *redis.Pool {
	if pool == nil {
		pool = NewPool(
			*redisAddress, *redisPassword, *redisMaxIdle, *redisMaxActive,
			time.Duration(*redisIdleTimeout)*time.Second,
		)
	}
	return pool
}

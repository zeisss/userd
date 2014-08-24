package eventstream

import (
	"github.com/garyburd/redigo/redis"
	"log"
)

func NewRedisEventStream(pool *redis.Pool, prefix string, pubsub bool) *redisEventStream {
	return &redisEventStream{prefix, pool, pubsub}
}

type redisEventStream struct {
	Prefix     string
	Pool       *redis.Pool
	UsePublish bool
}

func (stream *redisEventStream) channel(tag string) string {
	if stream.Prefix == "" {
		return tag
	}
	return stream.Prefix + "." + tag
}

func (stream *redisEventStream) Publish(tag string, data []byte) {
	con := stream.Pool.Get()
	defer con.Close()

	cmd := "RPUSH"
	channel := stream.channel(tag)
	msg := string(data)

	if stream.UsePublish {
		cmd = "PUBLISH"
	}

	_, err := con.Do(cmd, channel, msg)
	if err != nil {
		log.Fatalf("Failed to publish (%s): %s", channel, msg)
	}
}

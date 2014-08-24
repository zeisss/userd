package main

import (
	"./service"
	"./service/eventstream"
	"./service/hasher"
	"./service/idfactory"
	"./service/storage"

	httpcli "./http/cli"

	"flag"
	"os"
)

var (
	// Service/Logic
	authEmail = flag.Bool("auth-email", true, "Must the email adress be verified for an authentication to succeed.")

	// Backend Switches
	backendStorage = flag.String("storage", "memory", "Data storage: memory or etcd")

	// Backend - Hasher
	hasherBcryptCost = flag.Int("hasher-bcrypt-cost", hasher.BcryptDefaultCost, "The cost to apply when hashing new passwords.")

	// Backend - Storage
	/// Etcd
	storageEtcdPeer   = flag.String("storage-etcd-peer", "http://localhost:4001/", "The peer to connect to.")
	storageEtcdPrefix = flag.String("storage-etcd-prefix", "moinz.de/userd", "The path prefix to use with Etcd.")
	storageEtcdTtl    = flag.Uint64("storage-etcd-ttl", 365*24*60*60, "The TTL to use when creating entries in Etcd.")
)

func UserStorage() service.UserStorage {
	switch *backendStorage {
	case "redis":
		return storage.NewRedisStorage(RedisPool())
	case "etcd":
		return storage.NewEtcdStorage(*storageEtcdPeer, *storageEtcdPrefix, *storageEtcdTtl)
	case "memory":
		return storage.NewLocalStorage()
	default:
		panic("Unknown --storage value: " + *backendStorage)
	}
}

func IdFactory() service.IdFactory {
	return &idfactory.UUIDFactory{}
}

func PasswordHasher() service.PasswordHasher {
	return hasher.NewBcryptHasher(*hasherBcryptCost)
}

// ------------------------------------------------------------------------------
var (
	switchEventStream = flag.String("eventstream", "none", "Should events be logger? log or none")

	eventstreamRedisPrefix = flag.String("eventstream-redis-prefix", "", "A prefix to include into the queue/channel name")
	eventstreamRedisPubSub = flag.Bool("eventstream-redis-pubsub", false, "Use PUBLISH instead of RPUSH to send the message")

	eventstreamLogFile = flag.String("eventstream-log-file", "-", "Where to write the eventlog. - for stdout.")
	eventstreamLogMode = flag.Uint("eventstream-log-mode", 0600, "Mode to create logfile with - defaults to 0600.")

	eventstreamCoresUrl    = flag.String("eventstream-cores-url", "amqp://guest:guest@localhost", "An amqp url to connect to.")
	eventstreamCoresPrefix = flag.String("eventstream-cores-prefix", "", "A prefix to include into the routing key")
)

func EventStream() service.EventStream {
	switch *switchEventStream {
	case "redis":
		return eventstream.NewRedisEventStream(RedisPool(), *eventstreamRedisPrefix, *eventstreamRedisPubSub)
	case "cores":
		return eventstream.NewCoresAmqpEventLog(*eventstreamCoresUrl, *eventstreamCoresPrefix)
	case "log":
		return eventstream.NewFileLogEventStream(*eventstreamLogFile, os.FileMode(*eventstreamLogMode))
	case "none":
		return eventstream.NewNoneEventLog()
	default:
		panic("Unknown -eventstream value: " + *switchEventStream)
	}
}

func main() {
	flag.Parse()

	idFactory := IdFactory()
	hasher := PasswordHasher()
	userStorage := UserStorage()

	userService := service.UserService{
		service.Dependencies{idFactory, hasher, userStorage, EventStream()},
		service.Config{*authEmail},
	}

	handler := NewUserAPIHandler(&userService)
	httpcli.StartHttpInterface(handler)
}

package main

import (
	"./service"
	"./service/events"
	"./service/hasher"
	"./service/idfactory"
	"./service/storage"

	httputil "./http"

	"flag"
	"net/http"
	"os"
	"time"
)

var (
	// Service/Logic
	listenAddress  = flag.String("listen", "localhost:8080", "The address to listen on.")
	authEmail      = flag.Bool("auth-email", true, "Must the email adress be verified for an authentication to succeed.")
	wrapLogHandler = flag.Bool("log-requests", false, "Should requests be logged to stdout")

	// Frontend - HTTP
	httpsUse             = flag.Bool("https-enable", false, "Enable HTTPS listening in favor of HTTP.")
	httpsCertificateFile = flag.String("https-certificate", "server.cert", "The certificate to use for SSL.")
	httpsKeyFile         = flag.String("https-key", "server.key", "The keyfile to use for SSL.")

	// Backend Switches
	backendEventLog = flag.String("eventlog", "none", "Should events be logger? log or none")
	backendStorage  = flag.String("storage", "memory", "Data storage: memory or redis")

	// Backend - Hasher
	hasherBcryptCost = flag.Int("hasher-bcrypt-cost", hasher.BcryptDefaultCost, "The cost to apply when hashing new passwords.")

	// Backend - EventLog
	/// Log
	eventLogFile = flag.String("event-log-file", "-", "Where to write the eventlog. - for stdout.")
	eventLogMode = flag.Uint("event-log-mode", 0700, "Mode to create logfile with - defaults to 0700.")

	/// Cores
	eventCoresUrl = flag.String("event-cores-url", "amqp://guest:guest@localhost", "An amqp url to connect to.")

	// Backend - Storage
	/// Redis
	storageRedisAddress   = flag.String("storage-redis-address", ":6379", "The redis address to connect to host.")
	storageRedisMaxIdle   = flag.Int("storage-redis-max-idle", 3, "Maximum number of idle connections in the pool.")
	storageRedisMaxActive = flag.Int("storage-redis-max-active", 20, "Maximum number of active connections in the pool.")
	storageRedisTimeout   = flag.Int("storage-redis-pool-timeout", 240, "Seconds to keep idle connections in the pool.")
	storageRedisPassword  = flag.String("storage-redis-password", "", "A password to use for authorization.")
)

func UserStorage() service.UserStorage {
	switch *backendStorage {
	case "redis":
		return storage.NewRedisStorage(
			*storageRedisAddress,
			*storageRedisMaxIdle,
			*storageRedisMaxActive,
			time.Duration(*storageRedisTimeout)*time.Second,
			*storageRedisPassword,
		)
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

func EventLog() service.EventLog {
	switch *backendEventLog {
	case "cores":
		return events.NewCoresAmqpEventLog(*eventCoresUrl)
	case "log":
		var err error
		out := os.Stdout
		if *eventLogFile != "-" {
			out, err = os.OpenFile(*eventLogFile, os.O_WRONLY, os.FileMode(*eventLogMode))
			if err != nil {
				panic(err)
			}
		}
		return events.NewLogStreamEventLog(out)
	case "none":
		return events.NewNoneEventLog()
	default:
		panic("Unknown -eventlog value: " + *backendEventLog)
	}
}

func StartHttpInterface(userService *service.UserService) {
	handler := NewUserAPIHandler(userService)
	if *wrapLogHandler {
		handler = &httputil.RequestLogger{handler}
	}

	if *httpsUse {
		if err := http.ListenAndServeTLS(*listenAddress, *httpsCertificateFile, *httpsKeyFile, handler); err != nil {
			panic(err)
		}
	} else {
		if err := http.ListenAndServe(*listenAddress, handler); err != nil {
			panic(err)
		}
	}
}

func main() {
	flag.Parse()

	idFactory := IdFactory()
	hasher := PasswordHasher()
	userStorage := UserStorage()
	eventLog := EventLog()

	userService := service.UserService{
		service.Dependencies{idFactory, hasher, userStorage, eventLog},
		service.Config{*authEmail},
	}

	StartHttpInterface(&userService)
}

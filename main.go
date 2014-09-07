package main

import (
	"./middlewares"
	"./middlewares/v1"
	"./service"
	"./service/eventstream"
	"./service/hasher"
	"./service/idfactory"

	"./service/storage"

	httpcli "./http/cli"

	flag "github.com/ogier/pflag"
	metrics "github.com/rcrowley/go-metrics"

	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

// ------------------------------------------------------------------------------

var (
	// Backend Switches
	backendStorage         = flag.String("storage", "memory", "Data storage: memory, redis or etcd")
	storageEtcdPeers       = flag.String("storage-etcd-peers", "http://localhost:4001/", "The peers to connect to (comma separated).")
	storageEtcdPrefix      = flag.String("storage-etcd-prefix", "moinz.de/userd", "The path prefix to use with Etcd.")
	storageEtcdLogCURL     = flag.Bool("storage-etcd-log-curl", false, "Log calls to ETCD as curl commands to stdout.")
	storageEtcdLogFile     = flag.String("storage-etcd-log", "", "Filepath to write etcd debug log. Use - for stdout.")
	storageEtcdSyncCluster = flag.Bool("storage-etcd-sync-cluster", false, "Call SyncCluster initially to fetch all available nodes.")
	storageEtcdTtl         = flag.Uint64("storage-etcd-ttl", 0, "The TTL to use when creating entries in Etcd. 0 = no ttl")
)

func UserStorage() service.UserStorage {
	switch *backendStorage {
	case "redis":
		return storage.NewRedisStorage(RedisPool())
	case "etcd":
		var etcdLog *log.Logger

		switch *storageEtcdLogFile {
		case "":
			etcdLog = nil
		case "-":
			etcdLog = log.New(os.Stdout, "etcd", log.LstdFlags)
		default:
			out, err := os.OpenFile(*storageEtcdLogFile, os.O_WRONLY, os.FileMode(0700))
			if err != nil {
				panic(err)
			}
			etcdLog = log.New(out, "etcd", log.LstdFlags)
		}

		peers := strings.Split(*storageEtcdPeers, ",")
		return storage.NewEtcdStorage(peers, *storageEtcdPrefix, *storageEtcdTtl, *storageEtcdSyncCluster, *storageEtcdLogCURL, etcdLog)
	case "memory":
		return storage.NewLocalStorage()
	default:
		log.Fatalf("Unknown --storage value: %s", *backendStorage)
		return nil
	}
}

// ------------------------------------------------------------------------------

var (
	switchFactory = flag.String("idfactory", "uuid", "How to generate new IDs. uuid or seq")

	factorySeqFormat = flag.String("idfactory-seq-format", "user-%d", "The format when creating new IDs.")
)

func IdFactory() service.IdFactory {
	switch *switchFactory {
	case "uuid":
		return &idfactory.UUIDFactory{}
	case "seq":
		return idfactory.NewSequenceFactory(*factorySeqFormat)
	default:
		log.Fatalf("Unknown -idfactory value: %s", *switchFactory)
		return nil
	}

}

// ------------------------------------------------------------------------------

var (
	hasherBcryptCost = flag.Int("hasher-bcrypt-cost", hasher.BcryptDefaultCost, "The cost to apply when hashing new passwords.")
)

func PasswordHasher() service.PasswordHasher {
	return hasher.NewBcryptHasher(*hasherBcryptCost)
}

// ------------------------------------------------------------------------------

var (
	switchEventStreams = flag.String("eventstreams", "none", "Should events be logged? Use log, cores, redis or none")

	eventstreamRedisPrefix = flag.String("eventstream-redis-prefix", "", "A prefix to include into the queue/channel name")
	eventstreamRedisPubSub = flag.Bool("eventstream-redis-pubsub", false, "Use PUBLISH instead of RPUSH to send the message")

	eventstreamLogFile = flag.String("eventstream-log-file", "-", "Where to write the eventlog. - for stdout.")
	eventstreamLogMode = flag.Uint("eventstream-log-mode", 0600, "Mode to create logfile with - defaults to 0600.")

	eventstreamCoresUrl    = flag.String("eventstream-cores-url", "amqp://guest:guest@localhost", "An amqp url to connect to.")
	eventstreamCoresPrefix = flag.String("eventstream-cores-prefix", "", "A prefix to include into the routing key")
)

func EventStreams() *eventstream.Broadcaster {
	streamNames := strings.Split(*switchEventStreams, ",")
	broadcaster := eventstream.NewBroadcaster()

	if len(streamNames) == 0 {
		return broadcaster
	}

	for _, name := range streamNames {
		var newStream eventstream.Stream
		switch name {
		case "redis":
			newStream = eventstream.NewRedisEventStream(RedisPool(), *eventstreamRedisPrefix, *eventstreamRedisPubSub)
		case "cores":
			newStream = eventstream.NewCoresAmqpEventLog(*eventstreamCoresUrl, *eventstreamCoresPrefix)
		case "log":
			newStream = eventstream.NewFileLogEventStream(*eventstreamLogFile, os.FileMode(*eventstreamLogMode))
		case "none":
			newStream = eventstream.NewNoneEventLog()
		default:
			log.Fatalf("Unknown -eventstream value: %s", name)
		}

		broadcaster.AddStream(newStream)
	}
	return broadcaster
}

// ------------------------------------------------------------------------------

var (
	metricsCaptureDebugGCStats       = flag.Bool("metrics-capture-debug-stats", false, "Capture GC Debug stats in the metrics.")
	metricsCaptureDebugGCDuration    = flag.Int("metrics-capture-debug-duration", 60, "Duration between catching debug stats.")
	metricsCaptureRuntimeMemStats    = flag.Bool("metrics-capture-runtime-stats", true, "Capture Runtime Mem Stats.")
	metricsCaptureRuntimeMemDuration = flag.Int("metrics-capture-runtime-duration", 60, "Duration between catching runtime stats.")

	metricsGraphiteEnable = flag.Bool("metrics-graphite", false, "Write metrics to graphite")
	metricsGraphiteHost   = flag.String("metrics-graphite-host", "localhost:2003", "The host:port to connect to.")
	metricsGraphiteFlush  = flag.Int("metrics-graphite-flush", 10, "Seconds between flushes to graphite")
	metricsGraphitePrefix = flag.String("metrics-graphite-prefix", "userd", "A prefix for key sent to graphite")
)

func MetricsRegistry() metrics.Registry {
	r := metrics.DefaultRegistry

	if *metricsCaptureDebugGCStats {
		metrics.RegisterDebugGCStats(r)
		go metrics.CaptureDebugGCStats(r, time.Duration(*metricsCaptureDebugGCDuration)*time.Second)
	}
	if *metricsCaptureRuntimeMemStats {
		metrics.RegisterRuntimeMemStats(r)
		go metrics.CaptureRuntimeMemStats(r, time.Duration(*metricsCaptureRuntimeMemDuration)*time.Second)
	}

	if *metricsGraphiteEnable {
		addr, err := net.ResolveTCPAddr("tcp", *metricsGraphiteHost)
		if err != nil {
			panic(err)
		}
		go metrics.Graphite(r, time.Duration(*metricsGraphiteFlush)*time.Second, *metricsGraphitePrefix, addr)
	}

	return r
}

// ------------------------------------------------------------------------------

var (
	authEmail              = flag.Bool("auth-email", true, "Must the email adress be verified for an authentication to succeed.")
	eventCollectorMaxItems = flag.Int("feed-max-items", 1000, "Maximum items to keep in feed.")

	resetPasswordExpireTime = flag.Uint("expire-reset-password-token", 2*60, "How long can a resetPasswordToken be used (minutes)")
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	starter := httpcli.NewStarterFromFlagSet(flag.CommandLine)

	flag.Parse()

	MetricsRegistry() // Just calling, because it uses the DefaultRegistry

	dependencies := service.Dependencies{IdFactory(), PasswordHasher(), UserStorage(), EventStreams()}
	config := service.Config{*authEmail, *eventCollectorMaxItems, time.Duration(*resetPasswordExpireTime) * time.Minute}

	userService := service.NewUserService(config, dependencies)

	mux := http.NewServeMux()
	mux.Handle("/", middlewares.WelcomeHandler{})
	mux.Handle("/v1/", v1.NewUserAPIHandler(userService))
	starter.StartHttpInterface(mux)
}

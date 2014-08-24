package eventstream

import (
	cores "github.com/catalyst-zero/cores-go"

	"log"
	"os"
)

func NewCoresAmqpEventLog(amqpUrl, prefix string) *coresEventLog {
	bus, err := cores.NewAmqpEventBus(amqpUrl)
	if err != nil {
		panic(err)
	}
	return &coresEventLog{
		Bus:    bus,
		Prefix: prefix,
		Logger: log.New(os.Stderr, "[cores] ", log.LstdFlags),
	}
}

type coresEventLog struct {
	Bus    cores.EventBus
	Prefix string
	Logger *log.Logger
}

func (log *coresEventLog) withPrefix(tag string) string {
	if log.Prefix != "" {
		return log.Prefix + "." + tag
	}
	return tag
}

func (log *coresEventLog) Publish(tag string, data []byte) {
	if err := log.Bus.Publish(log.withPrefix(tag), data); err != nil {
		log.Logger.Fatalf("Failed to publish event to cores: %v", err)
	}
}

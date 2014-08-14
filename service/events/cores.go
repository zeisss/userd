package events

import (
	cores "github.com/catalyst-zero/cores-go"

	"log"
	"os"
)

func NewCoresAmqpEventLog(amqpUrl string) *coresEventLog {
	bus, err := cores.NewAmqpEventBus(amqpUrl)
	if err != nil {
		panic(err)
	}

	return &coresEventLog{
		Bus:    bus,
		Logger: log.New(os.Stderr, "[cores] ", log.LstdFlags),
	}
}

type coresEventLog struct {
	Bus    cores.EventBus
	Logger *log.Logger
}

func (log *coresEventLog) Log(tag string, data []byte) {
	if err := log.Bus.Publish(tag, data); err != nil {
		log.Logger.Fatalf("Failed to publish event to cores: %v", err)
	}
}

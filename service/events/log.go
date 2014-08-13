package events

import (
	"io"
	"log"
)

func NewLogStreamEventLog(writer io.Writer) *logStreamEventLog {
	logger := log.New(writer, "[userd] ", log.LstdFlags)
	return &logStreamEventLog{logger}
}

type logStreamEventLog struct {
	Logger *log.Logger
}

func (log *logStreamEventLog) Log(tag string, data []byte) {
	log.Logger.Printf("%s '%s'", tag, string(data))
}

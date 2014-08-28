package eventstream

import (
	"io"
	"log"
	"os"
)

func NewFileLogEventStream(filepath string, mode os.FileMode) *logEventStream {
	if filepath == "-" || filepath == "" {
		return NewLogEventStream(os.Stdout)
	} else {
		out, err := os.OpenFile(filepath, os.O_WRONLY, mode)
		if err != nil {
			panic(err)
		}
		return NewLogEventStream(out)
	}
}

func NewLogEventStream(out io.Writer) *logEventStream {
	logger := log.New(out, "[events] ", log.LstdFlags)
	return &logEventStream{logger}
}

type logEventStream struct {
	Logger *log.Logger
}

func (log *logEventStream) Publish(event string, data []byte) {
	log.Logger.Printf("publish %s: %s\n", event, string(data))
}

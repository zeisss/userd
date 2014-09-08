package microservice

import (
	"log"
	"os"
)

type ErrorReporter interface {
	Notify(opName string, input interface{}, err error)
}

func NewStderrErrorReporter() ErrorReporter {
	return &stdoutErrorReporter{
		Logger: log.New(os.Stderr, "ERROR ", log.LstdFlags),
	}
}

type stdoutErrorReporter struct {
	Logger *log.Logger
}

func (r *stdoutErrorReporter) Notify(opName string, input interface{}, err error) {
	r.Logger.Printf("%s (%+v) failed: %v", opName, input, err)
}

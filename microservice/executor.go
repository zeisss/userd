package microservice

import (
	metrics "github.com/rcrowley/go-metrics"

	"log"
	"time"
)

func NewExecutor(reporter ErrorReporter) *Executor {
	return &Executor{
		LogCalls:      true,
		ErrorReporter: reporter,
		Registry:      metrics.DefaultRegistry,

		operations: make(map[string]operation),
	}
}

type operation struct {
	Name    string
	Success metrics.Counter
	Failure metrics.Counter
	Timer   metrics.Timer
}

// TODO: Better design: start a collector in a go routine and let each operate execute independently
// and then send the result (success/failture + time) to the collector go routine.
type Executor struct {
	LogCalls      bool
	ErrorReporter ErrorReporter
	Registry      metrics.Registry

	operations map[string]operation
}

// execute executes the given `op` and returns it return value.
// If logging is enabled, the operation result and/or input will be logged.
// The call counter for the operation will be incremented and the execution time added to the metrics.
// If an error occurs, the error will also reported to the error reporter.
//
// No new errors should be returned. Anything the `op` returns, this method returns.
// The `output` is not modified.
func (executor *Executor) Execute(opName string, input interface{}, output interface{}, op func() error) error {
	if executor.LogCalls {
		log.Printf("call %s in  (%v)", opName, input)
	}

	// Execute
	now := time.Now()
	err := op()
	taken := time.Since(now)

	if executor.LogCalls {
		log.Printf("call %s out (%v, %v)", opName, output, err)
	}

	// Post processing
	executor.collectMetrics(opName, taken, err)
	executor.collectError(opName, input, err)

	return err
}

func (executor *Executor) collectError(opName string, input interface{}, err error) {
	if executor.ErrorReporter != nil && err != nil {
		executor.ErrorReporter.Notify(opName, input, err)
	}
}

func (executor *Executor) collectMetrics(opName string, dur time.Duration, err error) {
	m, found := executor.operations[opName]
	if !found {
		m = operation{
			Name:    opName,
			Success: metrics.GetOrRegisterCounter("service."+opName+".success", executor.Registry),
			Failure: metrics.GetOrRegisterCounter("service."+opName+".failure", executor.Registry),
			Timer:   metrics.GetOrRegisterTimer("service."+opName+".timer", executor.Registry),
		}
		executor.operations[opName] = m
	}

	if err != nil {
		m.Failure.Inc(1)
	} else {
		m.Success.Inc(1)
	}
	m.Timer.Update(dur)
}

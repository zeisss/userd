package service

import (
	metrics "github.com/rcrowley/go-metrics"

	"log"
	"time"
)

type errorReporter interface {
	Notify(opName string, input interface{}, err error)
}

func NewMetricExecutor() *MetricExecutor {
	return &MetricExecutor{
		LogCalls:      true,
		ErrorReporter: nil,
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
type MetricExecutor struct {
	LogCalls      bool
	ErrorReporter errorReporter
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
func (executor *MetricExecutor) execute(opName string, input interface{}, output interface{}, op func() error) error {
	if executor.LogCalls {
		log.Printf("call %s in  (%v)", opName, input)
	}

	// Execute
	now := time.Now()
	err := op()
	taken := time.Since(now)

	// Post processing
	executor.collectMetrics(opName, taken, err)

	if executor.LogCalls {
		log.Printf("call %s out (%v, %v)", opName, output, err)
	}

	if err != nil {
		if executor.ErrorReporter != nil {
			executor.ErrorReporter.Notify(opName, input, err)
		}
	} else {
		// TODO: Write to EventStream?
	}
	return err
}

func (executor *MetricExecutor) collectMetrics(opName string, dur time.Duration, err error) {
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

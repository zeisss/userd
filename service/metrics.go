package service

import (
	metrics "github.com/rcrowley/go-metrics"

	"log"
)

func NewMetricExecutor() *MetricExecutor {
	return &MetricExecutor{
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
	operations map[string]operation
}

// execute executes the given op and returns it return value.
// If logging is enabled, the operation result and/or input will be logged.
// The call counter for the operation will be incremented and the execution time added to the metrics.
// If an error occurs, the error will also reported to the error reporter.
//
// No new errors should be returned. Anything the `op` returns, this method returns.
// The `output` is not modified.
func (executor *MetricExecutor) execute(opName string, input interface{}, output interface{}, op func() error) error {
	log.Printf("call %s in  (%v)", opName, input)

	m, found := executor.operations[opName]
	if !found {
		m = operation{
			Name:    opName,
			Success: metrics.GetOrRegisterCounter("service."+opName+".success", metrics.DefaultRegistry),
			Failure: metrics.GetOrRegisterCounter("service."+opName+".failure", metrics.DefaultRegistry),
			Timer:   metrics.GetOrRegisterTimer("service."+opName+".timer", metrics.DefaultRegistry),
		}
		executor.operations[opName] = m
	}

	var err error
	m.Timer.Time(func() {
		err = op()
	})
	log.Printf("call %s out (%v, %v)", opName, output, err)
	if err != nil {
		// TODO: Report to error handler
		m.Failure.Inc(1)
	} else {

		// TODO: Write to EventStream?
		m.Success.Inc(1)
	}
	return err
}

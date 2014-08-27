package service

import (
	metrics "github.com/rcrowley/go-metrics"
)

func NewSuccessFailureCounter(prefix string) SuccessFailureCounter {
	success := metrics.GetOrRegisterCounter(prefix+".success", metrics.DefaultRegistry)
	failure := metrics.GetOrRegisterCounter(prefix+".failure", metrics.DefaultRegistry)

	return SuccessFailureCounter{success, failure}
}

type SuccessFailureCounter struct {
	success metrics.Counter
	failure metrics.Counter
}

func (c *SuccessFailureCounter) Success() {
	c.success.Inc(1)
}

func (c *SuccessFailureCounter) Failure() {
	c.failure.Inc(1)
}

func (c *SuccessFailureCounter) CountError(err error) error {
	if err != nil {
		c.Failure()
	} else {
		c.Success()
	}
	return err
}

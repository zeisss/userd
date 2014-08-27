package service

import (
	metrics "github.com/rcrowley/go-metrics"
)

func NewSuccessFailureMetric(prefix string) SuccessFailureMetric {
	success := metrics.GetOrRegisterCounter(prefix+".success", metrics.DefaultRegistry)
	failure := metrics.GetOrRegisterCounter(prefix+".failure", metrics.DefaultRegistry)
	timer := metrics.GetOrRegisterTimer(prefix+".timer", metrics.DefaultRegistry)

	return SuccessFailureMetric{success, failure, timer}
}

type SuccessFailureMetric struct {
	success metrics.Counter
	failure metrics.Counter
	timer   metrics.Timer
}

func (c *SuccessFailureMetric) Success() {
	c.success.Inc(1)
}

func (c *SuccessFailureMetric) Failure() {
	c.failure.Inc(1)
}

func (c *SuccessFailureMetric) CountError(err error) error {
	if err != nil {
		c.Failure()
	} else {
		c.Success()
	}
	return err
}

func (c *SuccessFailureMetric) Run(f func() error) error {
	var err error
	c.timer.Time(func() {
		err = f()
		c.CountError(err)
	})
	return err
}

func (c *SuccessFailureMetric) Time(f func()) {
	c.timer.Time(f)
}

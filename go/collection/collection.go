package collection

import "time"

// Metric is an interface containing a metric
type Metric interface {
	When() time.Time // when the metric was taken
}

// Collection contains a collection of Metrics
type Collection struct {
}

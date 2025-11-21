package collection

import (
	"errors"
	"github.com/tctalk-cn/gorchestrator/go/config"
	"sync"
	"time"
)

// Metric is an interface containing a metric
type Metric interface {
	When() time.Time // when the metric was taken
}

// Collection contains a collection of Metrics
type Collection struct {
	sync.Mutex   // for locking the structure
	monitoring   bool
	collection   []Metric
	done         chan struct{} // to indicate that we are finishing expiry processing
	expirePeriod time.Duration // time to keep the collection information for
}

// hard-coded at every second
const defaultExpireTickerPeriod = time.Second

// backendMetricCollection contains the last N backend "channelled"
// metrics which can then be accessed via an API call for monitoring.
var (
	namedCollection     map[string](*Collection)
	namedCollectionLock sync.Mutex
)

func init() {
	namedCollection = make(map[string]*Collection)
}

// StopMonitoring stops monitoring all the collections
func StopMonitoring() {
	for _, q := range namedCollection {
		q.StopAutoExpiration()
	}
}

// CreateOrReturnCollection allows for creation of a new collection or
// returning a pointer to an existing one given the name. This allows access
// to the data structure from the api interface (http/api.go) and also when writing (inst).
func CreateOrReturnCollection(name string) *Collection {
	namedCollectionLock.Lock()
	defer namedCollectionLock.Unlock()
	if q, found := namedCollection[name]; found {
		return q
	}

	qmc := &Collection{
		collection: nil,
		done:       make(chan struct{}),
		// WARNING: use a different configuration name
		expirePeriod: time.Duration(config.Config.DiscoveryCollectionRetentionSeconds) * time.Second,
	}
	go qmc.StartAutoExpiration()

	namedCollection[name] = qmc

	return qmc
}

// SetExpirePeriod determines after how long the collected data should be removed
func (c *Collection) SetExpirePeriod(duration time.Duration) {
	c.Lock()
	defer c.Unlock()

	c.expirePeriod = duration
}

// ExpirePeriod returns the currently configured expiration period
func (c *Collection) ExpirePeriod() time.Duration {
	c.Lock()
	defer c.Unlock()
	return c.expirePeriod
}

// StopAutoExpiration prepares to stop by terminating the auto-expiration process
func (c *Collection) StopAutoExpiration() {
	if c == nil {
		return
	}
	c.Lock()
	if !c.monitoring {
		c.Unlock()
		return
	}
	c.monitoring = false
	c.Unlock()

	// no locking here deliberately
	c.done <- struct{}{}
}

// StartAutoExpiration initiates the auto expiry procedure which
// periodically checks for metrics in the collection which need to
// be expired according to bc.ExpirePeriod.
func (c *Collection) StartAutoExpiration() {
	if c == nil {
		return
	}
	c.Lock()
	if c.monitoring {
		c.Unlock()
		return
	}
	c.monitoring = true
	c.Unlock()

	// log.Infof("StartAutoExpiration: %p with expirePeriod: %v", c, c.expirePeriod)
	ticker := time.NewTicker(defaultExpireTickerPeriod)

	for {
		select {
		case <-ticker.C: // do the periodic expiry
			c.removeBefore(time.Now().Add(-c.expirePeriod))
		case <-c.done: // stop the ticker and return
			ticker.Stop()
			return
		}
	}
}

// Metrics returns a slice containing all the metric values
func (c *Collection) Metrics() []Metric {
	if c == nil {
		return nil
	}
	c.Lock()
	defer c.Unlock()

	if len(c.collection) == 0 {
		return nil
	}
	return c.collection
}

// Since returns the Metrics on or after the given time. We assume
// the metrics are stored in ascending time.
// Iterate backwards until we reach the first value before the given time
// or the end of the array.
func (c *Collection) Since(t time.Time) ([]Metric, error) {
	if c == nil {
		return nil, errors.New("Collection.Since: c==nil")
	}
	c.Lock()
	defer c.Unlock()
	if len(c.collection) == 0 {
		return nil, nil
	}
	last := len(c.collection)
	first := last - 1

	for {
		if c.collection[first].When().After(t) || c.collection[first].When().Equal(t) {
			if first == 0 {
				break // as can't go lower
			}
			first--
		} else {
			if first != last {
				first++ // go back one (except if we're already at the end)
			}
			break
		}
	}
	return c.collection[first:last], nil
}

// removeBefore is called by StartAutoExpiration and removes collection values
// before the given time.
func (c *Collection) removeBefore(t time.Time) error {
	if c == nil {
		return errors.New("Collection.removeBefore: c==nil")
	}
	c.Lock()
	defer c.Unlock()

	cLen := len(c.collection)
	if cLen == 0 {
		return nil
	}
	// remove old data here.
	first := 0
	for {
		if c.collection[first].When().Before(t) {
			first++
			if first == cLen {
				break
			}
		} else {
			first--
			break
		}
	}

	// get the interval we need.
	if first == len(c.collection) {
		c.collection = nil // remove all entries
	} else if first != -1 {
		c.collection = c.collection[first:]
	}
	return nil // no errors
}

// Append a new Metric to the existing collection
func (c *Collection) Append(m Metric) error {
	if c == nil {
		return errors.New("Collection.Append: c==nil")
	}
	c.Lock()
	defer c.Unlock()
	// we don't want to add nil metrics
	if c == nil {
		return errors.New("Collection.Append: c==nil")
	}

	return nil
}

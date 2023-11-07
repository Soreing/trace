package trace

import (
	"time"

	"github.com/Soreing/grand"
)

// Configuration is a collection of options that apply to the client.
type Configuration struct {
	rand       Random
	batchTime  time.Duration
	batchCount int
}

// newConfiguration creates default configs and applies options
func newConfiguration(opts []Option) (*Configuration, error) {
	cfg := &Configuration{
		batchTime:  0,
		batchCount: 0,
	}

	for _, opt := range opts {
		opt.Configure(cfg)
	}

	if cfg.rand == nil {
		src, err := grand.NewSource()
		if err != nil {
			return nil, err
		}
		cfg.rand = grand.New(src)
	}

	return cfg, nil
}

// Option defines objects that can change a Configuration.
type Option interface {
	Configure(c *Configuration)
}

// UseRandomizer creates an option for setting the tracer's random generator.
func UseRandomizer(rand Random) Option {
	return &randOption{
		rand: rand,
	}
}

// UseBatching creates an option for batching spans before dispatching them.
func UseBatching(maxTime time.Duration, maxCount int) Option {
	return &batchOption{
		batchTime:  maxTime,
		batchCount: maxCount,
	}
}

type randOption struct {
	rand Random
}

func (o *randOption) Configure(c *Configuration) {
	c.rand = o.rand
}

type batchOption struct {
	batchTime  time.Duration
	batchCount int
}

func (o *batchOption) Configure(c *Configuration) {
	c.batchTime = o.batchTime
	c.batchCount = o.batchCount
}

package goTx

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

type RetryOptions struct {
	MaxRetries int
	Backoff    Backoff

	UnrecoverableErrors []error
}

type Backoff interface {
	NextInterval() time.Duration
}

type ConstantBackoff struct {
	Interval time.Duration
}

func (b *ConstantBackoff) NextInterval() time.Duration {
	return b.Interval
}

type ExponentialBackoff struct {
	InitialInterval time.Duration
	MaxInterval     time.Duration
	Multiplier      float64
	RandomFactor    float64
	CurrentInterval time.Duration
}

func (b *ExponentialBackoff) NextInterval() time.Duration {
	if b.CurrentInterval == 0 {
		b.CurrentInterval = b.InitialInterval
		return b.CurrentInterval
	}
	next := time.Duration(float64(b.CurrentInterval) * b.Multiplier)
	if b.RandomFactor > 0 {
		jitter := (2*rand.Float64() - 1) * b.RandomFactor
		next = time.Duration(float64(next) * (1 + jitter))
	}
	if next > b.MaxInterval {
		next = b.MaxInterval
	}
	b.CurrentInterval = next
	return next
}

func Retry(fn func() error, options RetryOptions) error {
	var err error
	for i := 0; i < options.MaxRetries; i++ {
		if err = fn(); err == nil {
			return nil
		} else if options.UnrecoverableErrors != nil && len(options.UnrecoverableErrors) > 0 {
			for _, e := range options.UnrecoverableErrors {
				if errors.Is(err, e) {
					return fmt.Errorf("unrecoverable error: %v", err)
				}
			}
		}
		time.Sleep(options.Backoff.NextInterval())
	}
	return fmt.Errorf("error after %d retries: %v", options.MaxRetries, err)
}

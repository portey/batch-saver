package storage

import (
	"context"

	"github.com/portey/batch-saver/models"
)

type sinker interface {
	Sink(cxt context.Context, events []models.Event) error
}

type RateLimiter struct {
	sinker sinker
	ch     chan struct{}
}

func NewRateLimiter(size int, sinker sinker) *RateLimiter {
	return &RateLimiter{
		sinker: sinker,
		ch:     make(chan struct{}, size),
	}
}

func (r *RateLimiter) Sink(ctx context.Context, events []models.Event) error {
	<-r.ch
	err := r.sinker.Sink(ctx, events)
	r.ch <- struct{}{}
	return err
}

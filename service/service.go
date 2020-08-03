package service

import (
	"context"
	"time"

	"github.com/portey/batch-saver/models"
	"github.com/portey/batch-saver/service/batching"
)

type Config struct {
	NumberOfWorkers   int
	BatchMaxSize      int
	BatchFlushTimeout time.Duration
}

type Service struct {
	sink chan<- models.Event
}

func New(ctx context.Context, cfg Config, sinker batching.Sinker) *Service {
	store := batching.NewBatchStore(ctx, sinker, cfg.BatchMaxSize, cfg.BatchFlushTimeout)
	sink := make(chan models.Event)

	for i := 0; i < cfg.NumberOfWorkers; i++ {
		batching.NewWorker(i, sink, store)
	}

	go func() {
		<-ctx.Done()
		close(sink)
	}()

	return &Service{
		sink: sink,
	}
}

func (s *Service) Sink(event models.Event) error {
	s.sink <- event
	return nil
}

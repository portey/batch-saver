package batching

import (
	"context"
	"sync"
	"time"

	"github.com/portey/batch-saver/models"
	log "github.com/sirupsen/logrus"
)

type Sinker interface {
	Sink(context.Context, []models.Event) error
}

type BatchStore struct {
	ctx          context.Context
	sinker       Sinker
	batches      *sync.Map
	maxSize      int
	flushTimeout time.Duration
}

func NewBatchStore(ctx context.Context, sinker Sinker, maxSize int, flushTimeout time.Duration) *BatchStore {
	batches := sync.Map{}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Minute):
				batches.Range(func(k, v interface{}) bool {
					if v, ok := v.(*Batch); ok && len(v.items) == 0 {
						v.checkDie <- struct{}{}
					}
					return true
				})
			}
		}
	}()

	return &BatchStore{
		ctx:          ctx,
		sinker:       sinker,
		batches:      &batches,
		maxSize:      maxSize,
		flushTimeout: flushTimeout,
	}
}

func (s *BatchStore) GetByGroup(id string) *Batch {
	b, ok := s.batches.Load(id)
	if ok {
		return b.(*Batch)
	}

	newB := newBatch(s.ctx, s.sinker, s.maxSize, s.flushTimeout)
	s.batches.Store(id, newB)

	return newB
}

type Batch struct {
	items    []models.Event
	checkDie chan struct{}
	ch       chan<- models.Event
}

func newBatch(ctx context.Context, sinker Sinker, maxSize int, flushTimeout time.Duration) *Batch { // nolint:gocyclo
	ch := make(chan models.Event)
	checkDie := make(chan struct{})
	items := make([]models.Event, 0, maxSize)

	timer := time.NewTimer(0)
	var timerCh <-chan time.Time

	flush := func() {
		log.Tracef("batching: flushing Batch of size %d", len(items))
		if len(items) == 0 {
			return
		}

		if err := sinker.Sink(ctx, items); err != nil {
			log.WithError(err).Error("batching: could't sink Batch")
			// todo implement backoff strategy here
		}

		items = items[:0]
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				close(ch)
				close(checkDie)
				flush()
				return
			case <-checkDie:
				if len(items) != 0 {
					continue
				}
				log.Trace("batching: cleaning batch")
				close(ch)
				close(checkDie)
				return
			case e, ok := <-ch:
				if !ok {
					return
				}

				items = append(items, e)
				if len(items) == maxSize {
					if !timer.Stop() {
						<-timer.C
					}
					flush()
					continue
				}

				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(flushTimeout)
				timerCh = timer.C
			case <-timerCh:
				timer.Stop()
				flush()
			}
		}
	}()

	return &Batch{items: items, ch: ch, checkDie: checkDie}
}

func (b *Batch) Add(event models.Event) {
	b.ch <- event
}

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
	mx           sync.Mutex
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
	s.mx.Lock()
	defer s.mx.Unlock()
	b, ok := s.batches.Load(id)
	if ok {
		return b.(*Batch)
	}

	newB := newBatch(s.ctx, s.sinker, s.maxSize, s.flushTimeout, func() {
		s.batches.Delete(id)
	})
	s.batches.Store(id, newB)

	return newB
}

type Batch struct {
	items    []models.Event
	checkDie chan struct{}
	ch       chan<- models.Event
}

func newBatch(ctx context.Context, sinker Sinker, maxSize int, flushTimeout time.Duration, cleanup func()) *Batch { // nolint:gocyclo
	ch := make(chan models.Event)
	checkDie := make(chan struct{})
	items := make([]models.Event, 0, maxSize)

	var timerCh <-chan time.Time

	flush := func() {
		log.Tracef("batching: flushing batch with size %d", len(items))
		if len(items) == 0 {
			return
		}

		if err := sinker.Sink(ctx, items); err != nil {
			log.WithError(err).Error("batching: could't sink batch")
			// todo implement backoff strategy here
		}

		items = items[:0]
	}

	go func() {
		defer cleanup()

		for {
			select {
			case <-ctx.Done():
				close(ch)
				close(checkDie)
				log.Trace("batching: context done", ctx.Err().Error())
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

				log.Tracef("batching: accepted message %s", e.ID)

				items = append(items, e)
				if len(items) == maxSize {
					log.Trace("batching: flushing batch by size")
					flush()
					continue
				}

				timerCh = time.After(flushTimeout)
			case _, ok := <-timerCh:
				if !ok {
					continue
				}
				log.Trace("batching: flushing batch by timeout")
				flush()
			}
		}
	}()

	return &Batch{items: items, ch: ch, checkDie: checkDie}
}

func (b *Batch) Add(event models.Event) {
	b.ch <- event
}

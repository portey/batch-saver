package batching

import (
	"github.com/portey/batch-saver/models"
	log "github.com/sirupsen/logrus"
)

func NewWorker(index int, in <-chan models.Event, store *BatchStore) {
	go doWork(index, in, store)
	log.Tracef("worker: worker %d started", index)
}

func doWork(index int, in <-chan models.Event, store *BatchStore) {
	for e := range in {
		log.Tracef("worker: worker %d received message id=%s", index, e.ID)

		store.GetByGroup(e.GroupID).Add(e)
	}
}

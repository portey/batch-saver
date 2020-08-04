package service

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/portey/batch-saver/models"
	"github.com/portey/batch-saver/service/batching/mock"
	"github.com/stretchr/testify/assert"
)

func TestService_TimeoutFlush(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	event := models.Event{
		ID:      "id",
		GroupID: "group_id",
		Data:    []byte("data"),
	}

	sinker := mock.NewMockSinker(ctrl)
	sinker.EXPECT().
		Sink(gomock.Eq(ctx), gomock.Eq([]models.Event{event})).
		Return(nil)

	srv := New(ctx, Config{
		NumberOfWorkers:   2,
		BatchMaxSize:      3,
		BatchFlushTimeout: time.Second,
	}, sinker)

	err := srv.Sink(event)
	assert.Nil(t, err)
	time.Sleep(2 * time.Second)
}

func TestService_SizeFlush(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	event := models.Event{
		ID:      "id",
		GroupID: "group_id",
		Data:    []byte("data"),
	}

	sinker := mock.NewMockSinker(ctrl)
	sinker.EXPECT().
		Sink(gomock.Eq(ctx), gomock.Eq([]models.Event{event, event, event})).
		Return(nil)

	srv := New(ctx, Config{
		NumberOfWorkers:   2,
		BatchMaxSize:      3,
		BatchFlushTimeout: time.Minute,
	}, sinker)

	err := srv.Sink(event)
	assert.Nil(t, err)
	err = srv.Sink(event)
	assert.Nil(t, err)
	err = srv.Sink(event)
	assert.Nil(t, err)

	time.Sleep(time.Second)
}

func TestService_Grouping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	event := models.Event{ID: "id", GroupID: "group_id", Data: []byte("data")}
	event2 := models.Event{ID: "id", GroupID: "group_id2", Data: []byte("data")}

	sinker := mock.NewMockSinker(ctrl)
	sinker.EXPECT().
		Sink(gomock.Eq(ctx), gomock.Any()).
		Times(2).
		DoAndReturn(func(_ context.Context, events []models.Event) error {
			assert.Len(t, events, 1)
			return nil
		})

	srv := New(ctx, Config{
		NumberOfWorkers:   2,
		BatchMaxSize:      3,
		BatchFlushTimeout: time.Second,
	}, sinker)

	err := srv.Sink(event)
	assert.Nil(t, err)
	err = srv.Sink(event2)
	assert.Nil(t, err)

	time.Sleep(2 * time.Second)
}

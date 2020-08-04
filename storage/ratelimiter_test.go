package storage

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/portey/batch-saver/models"
	"github.com/portey/batch-saver/storage/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewRateLimiter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	events := []models.Event{
		{
			ID:      "id",
			GroupID: "group_id",
			Data:    []byte("data"),
		},
	}

	sink := mock.NewMocksinker(ctrl)
	sink.EXPECT().
		Sink(gomock.Eq(ctx), gomock.Eq(events)).
		Times(1).
		Return(nil)

	limiter := NewRateLimiter(1, sink)
	err := limiter.Sink(ctx, events)
	assert.Nil(t, err)
}

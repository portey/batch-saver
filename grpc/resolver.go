package grpc

import (
	"io"
	"sync"

	"github.com/portey/batch-saver/gen/api"
	"github.com/portey/batch-saver/models"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type service interface {
	Sink(event models.Event) error
}

type resolver struct {
	service    service
	eventsPool sync.Pool
}

func newResolver(service service) *resolver {
	return &resolver{
		service: service,
		eventsPool: sync.Pool{
			New: func() interface{} {
				return &models.Event{}
			},
		},
	}
}

func (r *resolver) SaveEvent(stream api.BatchSaver_SaveEventServer) error {
loop:
	for {
		select {
		case <-stream.Context().Done():
		default:
			req, err := stream.Recv()
			if err == io.EOF {
				break loop
			}
			if err != nil {
				return status.Error(codes.Internal, err.Error())
			}

			e := r.eventsPool.Get().(*models.Event)
			e.ID = req.GetId()
			e.GroupID = req.GetGroupId()
			e.Data = req.GetData()

			err = r.service.Sink(*e)
			r.eventsPool.Put(e)

			if err != nil {
				log.WithError(err).Error("grpc: sink event error")
				return status.Error(codes.Internal, err.Error())
			}
		}
	}

	return stream.SendAndClose(&api.SaveEventResponse{})
}

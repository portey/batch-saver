package grpc

import "github.com/portey/batch-saver/gen/api"

type service interface {
}

type resolver struct {
	service service
}

func newResolver(service service) *resolver {
	return &resolver{
		service: service,
	}
}

func (r *resolver) SaveEvent(request api.BatchSaver_SaveEventServer) error {
	// todo implement logic here
	panic("not implemented")
}

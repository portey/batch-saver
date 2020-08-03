package service

import "github.com/portey/batch-saver/models"

type Service struct {
}

func New() *Service {
	return &Service{}
}

func (s *Service) Sink(event models.Event) error {

}

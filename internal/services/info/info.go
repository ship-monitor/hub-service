package info

import "github.com/google/uuid"

type Service struct {
	id uuid.UUID
}

func New() *Service {
	return &Service{
		id: uuid.MustParse("3dea52fe-4fa8-43cf-9ea1-d54c7b74d467"),
	}
}

func (s *Service) ID() uuid.UUID {
	return s.id
}

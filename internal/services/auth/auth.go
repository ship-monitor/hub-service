package auth

import (
	"context"

	"github.com/google/uuid"
)

type Info interface {
	ID() uuid.UUID
}

type Service struct {
	info Info
}

func New(
	info Info,
) *Service {
	return &Service{
		info: info,
	}
}

func (s *Service) GetAuthHeader(ctx context.Context) (string, error) {
	return s.info.ID().String(), nil
}

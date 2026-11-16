package types

import "github.com/google/uuid"

type Sensor struct {
	ID   uuid.UUID
	Name string
}

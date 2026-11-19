package types

import "github.com/google/uuid"

type Sensor struct {
	ID           uuid.UUID
	Name         string
	Manufacturer string
	Model        string
	Revision     string
	Software     string
}

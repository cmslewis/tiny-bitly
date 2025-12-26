package model

import "time"

type Entity struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	DeletedAt time.Time `json:"deletedAt"`
}

func (e Entity) IsDeleted() bool {
	return !e.DeletedAt.IsZero()
}

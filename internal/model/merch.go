package model

import "time"

type Merch struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	Price     int       `json:"price"`
	IsSelling bool      `json:"is_selling"`
	CreatedAt time.Time `json:"created_at"`
}

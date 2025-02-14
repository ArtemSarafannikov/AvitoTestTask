package model

import "time"

type User struct {
	Id        string    `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	Balance   int       `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}

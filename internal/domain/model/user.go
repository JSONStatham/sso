package model

import "time"

type User struct {
	ID        int
	Email     string
	Password  []byte
	CreatedAt time.Time
}

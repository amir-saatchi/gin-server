package models

import "time"

type Log struct {
	ID        uint   `gorm:"primarykey"`
	Message   string `gorm:"not null"`
	CreatedAt time.Time
}
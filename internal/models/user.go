package models

import (
    "time"
)

// User represents a user in the database
type User struct {
    ID        uint       `gorm:"primarykey"`
    Name      string     `gorm:"not null"`
    Email     string     `gorm:"unique;not null"`
    CreatedAt time.Time
    UpdatedAt time.Time
}
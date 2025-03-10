package models

import (
    "time"
    "github.com/google/uuid"
)

// Enums for Role and Profession
type Role string

const (
    AdminRole    Role = "ADMIN"
    CompanyRole  Role = "COMPANY"
    UserRole     Role = "USER"
)

type Profession string

const (
    AdminProf   Profession = "ADMIN"
    AnalystProf Profession = "ANALYST"
    ManagerProf Profession = "MANAGER"
    HeadProf    Profession = "HEAD"
)

type ProjectStatus string

const (
    InProgress ProjectStatus = "INPROGRESS"
    DoneStatus ProjectStatus = "DONE"
)

// User Model (Simplified for permissions/ownership)
type User struct {
    ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    Email      string    `gorm:"unique;not null;index;size:255" validate:"required,email"`
    CompanyId  uuid.UUID `gorm:"not null"` // Assume managed externally
    Role       Role      `gorm:"type:varchar(20);not null"`
    Profession Profession `gorm:"type:varchar(20);not null"`

    CreatedAt time.Time `gorm:"autoCreateTime"` // Audit field
    UpdatedAt time.Time `gorm:"autoUpdateTime"` // Audit field

    // Relationships
    Projects    []Project `gorm:"many2many:user_projects;"`
    OwnedProjects []Project `gorm:"foreignkey:OwnerId;"`
}

// Project Model
type Project struct {
    ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    CompanyId   uuid.UUID `gorm:"not null"` // Assume managed externally
    Name        string    `gorm:"not null;unique;size:255"`
    Description string    `gorm:"not null"`
    Status      ProjectStatus `gorm:"type:varchar(20);not null"`
    OwnerId     uuid.UUID `gorm:"not null"`
    Owner       User      `gorm:"foreignKey:OwnerId"`

    // Permissions (Many-to-Many)
    Viewers []User `gorm:"many2many:viewers;"`
    Editors []User `gorm:"many2many:editors;"`

    CreatedAt time.Time `gorm:"autoCreateTime"`
    UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// Join Table for Permissions (if needed explicitly)
type UserProject struct {
    UserID   uuid.UUID `gorm:"primary_key"`
    ProjectID uuid.UUID `gorm:"primary_key"`
}
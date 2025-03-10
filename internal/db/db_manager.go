package db

import (
	"fmt"
	"sync"

	"log"

	"github.com/amir-saatchi/rest-api/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DBManager manages multiple database connections
type DBManager struct {
    mu       sync.RWMutex
    dbs      map[string]*gorm.DB // Map company_id to DB instances
    dsnMap   map[string]string   // Map company_id to DSN strings
}

// NewDBManager initializes the DBManager
func NewDBManager(dsnMap map[string]string) *DBManager {
    return &DBManager{
        dbs:    make(map[string]*gorm.DB),
        dsnMap: dsnMap,
    }
}

// GetDB retrieves or creates a database connection for the given company_id
func (m *DBManager) GetDB(companyID string) (*gorm.DB, error) {
    m.mu.RLock()
    db, exists := m.dbs[companyID]
    m.mu.RUnlock()

    if exists && db != nil {
        return db, nil
    }

    // If the connection doesn't exist, create it
    m.mu.Lock()
    defer m.mu.Unlock()

    dsn, ok := m.dsnMap[companyID]
    if !ok {
        return nil, fmt.Errorf("no database configured for company_id: %s", companyID)
    }

    newDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Printf("Failed to connect to database for company_id: %s", companyID)
        return nil, err
    }

    // Auto-migrate schema (for development only)
    newDB.AutoMigrate(&models.Project{}, &models.User{})

    m.dbs[companyID] = newDB
    return newDB, nil
}
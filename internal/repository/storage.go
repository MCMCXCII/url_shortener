package repository

import (
	"database/sql"
	"fmt"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (p *PostgresRepository) Ping() error {
	if p.db == nil {
		return fmt.Errorf("db is nil")
	}
	return p.db.Ping()
}

func (p *PostgresRepository) Save(id, url string)          {}
func (p *PostgresRepository) Get(id string) (string, bool) { return "", true }
func (p *PostgresRepository) Load() error

package repository

import (
	"database/sql"
	"fmt"

	"github.com/MCMCXCII/url_shortener/internal/logger"
	"go.uber.org/zap"
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

func (p *PostgresRepository) Save(id, url string) error {
	_, err := p.db.Exec(`INSERT INTO urls (short, original) VALUES ($1, $2)`, id, url)
	return err
}
func (p *PostgresRepository) Get(id string) (string, bool) {
	var original string
	or := p.db.QueryRow(
		`SELECT original FROM urls WHERE short = $1`,
		id)
	err := or.Scan(&original)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", false
		}
		logger.Log.Debug("error select from db",
			zap.String("original", original))
	}
	return original, true
}

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

func (p *PostgresRepository) SaveBatch(items []BatchItem) error {
	if len(items) == 0 {
		return nil
	}

	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO urls (short, original) VALUES ($1, $2)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, item := range items {
		if _, err := stmt.Exec(item.ID, item.URL); err != nil {
			return err
		}
	}
	return tx.Commit()
}

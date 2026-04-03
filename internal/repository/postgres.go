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
	res, err := p.db.Exec(`
	INSERT INTO urls (short, original)
	VALUES ($1, $2)
	ON CONFLICT (original) DO NOTHING`, id, url)

	if err != nil {
		return fmt.Errorf("insert url: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return ErrOriginalURLExists
	}
	return nil
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

func (p *PostgresRepository) GetByOriginal(original string) (string, bool) {
	var short string

	row := p.db.QueryRow(`
	SELECT short
	FROM urls
	WHERE original == $1
	`, original)

	err := row.Scan(&short)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", false
		}
		logger.Log.Debug("error select by original", zap.String("original", original))
		return "", false
	}
	return short, true
}

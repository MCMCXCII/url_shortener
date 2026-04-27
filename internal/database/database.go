package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/MCMCXCII/url_shortener/internal/logger"
	"go.uber.org/zap"
)

// DB - интерфейс подключения к базе данных.
type DB interface {
	QueryRow(query string, args ...any) *sql.Row
	Query(query string, args ...any) (*sql.Rows, error)
	Exec(query string, args ...any) (sql.Result, error)
	Ping() error
	Close() error
	Begin() (*sql.Tx, error)
}

// db - конкретная реализация, скрыта от внешних пакетов.
type db struct {
	conn *sql.DB
}

func New(dsn string) (DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("dsn is empty")
	}

	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	conn.SetMaxOpenConns(10)
	conn.SetConnMaxLifetime(time.Hour)
	conn.SetConnMaxIdleTime(10)

	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	if err := migrate(conn); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to migrate")
	}

	logger.Log.Info("connected to database", zap.String("dsn", dsn))
	return &db{conn: conn}, nil
}

func (d *db) QueryRow(query string, args ...any) *sql.Row {
	return d.conn.QueryRow(query, args...)
}

func (d *db) Query(query string, args ...any) (*sql.Rows, error) {
	return d.conn.Query(query, args...)
}

func (d *db) Exec(query string, args ...any) (sql.Result, error) {
	return d.conn.Exec(query, args...)
}

func (d *db) Ping() error {
	return d.conn.Ping()
}

func (d *db) Close() error {
	logger.Log.Info("closing database connection")
	return d.conn.Close()
}

func migrate(conn *sql.DB) error {
	_, err := conn.Exec(`CREATE TABLE IF NOT EXISTS urls (
		id       SERIAL PRIMARY KEY,
		short    TEXT NOT NULL,
		original TEXT UNIQUE NOT NULL
	)`)
	return err
}

func (d *db) Begin() (*sql.Tx, error) {
	return d.conn.Begin()
}

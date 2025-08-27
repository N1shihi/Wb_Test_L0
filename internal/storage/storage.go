package storage

import (
	"Wb_Test_L0/internal/config"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

func New(cfg *config.Config) (*sql.DB, error) {
	conStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Name, cfg.DB.SslMode)
	db, err := sql.Open("postgres", conStr)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return db, nil
}

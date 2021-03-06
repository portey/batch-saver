package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/mattes/migrate"
	"github.com/portey/batch-saver/models"

	_ "github.com/mattes/migrate/database/postgres"
	_ "github.com/mattes/migrate/source/file"
)

type (
	Config struct {
		Host     string
		Port     int
		Db       string
		User     string
		Password string
		Ssl      bool
	}

	Storage struct {
		db *sqlx.DB
	}
)

func New(cfg Config) (*Storage, error) {
	db, err := sqlx.Connect("postgres", cfg.address())
	if err != nil {
		return nil, err
	}

	m, err := migrate.New("file://storage/postgres/migrations", cfg.address())
	if err != nil {
		return nil, err
	}
	defer m.Close()

	if err = m.Up(); err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

func (r *Storage) HealthCheck() error {
	return r.db.Ping()
}

func (config Config) address() string {
	address := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=", config.User, config.Password, config.Host, config.Port, config.Db)

	if !config.Ssl {
		address += "disable"
	}

	return address
}

func (r *Storage) Sink(ctx context.Context, events []models.Event) error {
	tx := r.db.MustBeginTx(ctx, nil)

	for _, event := range events {
		if _, err := tx.NamedExecContext(
			ctx,
			"INSERT INTO events(id, group_id, data) VALUES (:id, :group_id, :data) ON CONFLICT DO NOTHING",
			map[string]interface{}{
				"id":       event.ID,
				"group_id": event.GroupID,
				"data":     event.Data,
			}); err != nil {
			return tx.Rollback()
		}
	}

	return tx.Commit()
}

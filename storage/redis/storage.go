package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/portey/batch-saver/models"
)

type Config struct {
	Addr     string
	PoolSize int
}

type Storage struct {
	client *redis.Client
}

func New(ctx context.Context, cfg Config) (*Storage, error) {
	opt, err := redis.ParseURL(cfg.Addr)
	if err != nil {
		return nil, err
	}

	opt.PoolSize = cfg.PoolSize
	opt.Username = ""
	client := redis.NewClient(opt)

	err = client.Ping(ctx).Err()
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		err := client.Close()
		if err != nil {
			return
		}
	}()

	return &Storage{client: client}, nil
}

func (s *Storage) Sink(ctx context.Context, events []models.Event) error {
	if len(events) == 0 {
		return nil
	}

	items := make([]interface{}, len(events))
	for i := range events {
		items[i] = events[i]
	}

	return s.client.SAdd(ctx, "group_"+events[0].GroupID, items...).Err()
}

func (s *Storage) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	return s.client.Ping(ctx).Err()
}

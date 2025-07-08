package repo

import (
	"context"
	"fmt"
	"gateway/internal/config"
	"github.com/jackc/pgx/v5"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

// Слой репозитория, здесь должны быть все методы, связанные с базой данных

type repository struct {
	pool *pgxpool.Pool
}

// Repository - интерфейс с методом создания задачи
type Repository interface {
	CreateTask(ctx context.Context, task Task) (int, error)
	GetTask(ctx context.Context, id string) (*Task, error)
	CreateUser(ctx context.Context, username, passwordHash string) error
	GetUser(ctx context.Context, username string) (*User, error)
}

// NewRepository - создание нового экземпляра репозитория с подключением к PostgreSQL
func NewRepository(ctx context.Context, cfg config.PostgreSQL) (Repository, error) {
	// Формируем строку подключения
	connString := fmt.Sprintf(
		`user=%s password=%s host=%s port=%d dbname=%s sslmode=%s 
        pool_max_conns=%d pool_max_conn_lifetime=%s pool_max_conn_idle_time=%s`,
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
		cfg.PoolMaxConns,
		cfg.PoolMaxConnLifetime.String(),
		cfg.PoolMaxConnIdleTime.String(),
	)

	// Парсим конфигурацию подключения
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse PostgreSQL config")
	}

	// Оптимизация выполнения запросов (кеширование запросов)
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe

	// Создаём пул соединений с базой данных
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create PostgreSQL connection pool")
	}

	return &repository{pool}, nil
}

// CreateTask - вставка новой задачи в таблицу tasks
func (r *repository) CreateTask(ctx context.Context, task Task) (int, error) {
	var id int
	err := r.pool.QueryRow(ctx, insertTaskQuery, task.Title, task.Description).Scan(&id)
	if err != nil {
		return 0, errors.Wrap(err, "failed to insert task")
	}
	return id, nil
}

// GetTask - получение задачи из таблицы task по task.id
func (r *repository) GetTask(ctx context.Context, id string) (*Task, error) {
	task := Task{}
	err := r.pool.QueryRow(ctx, getTaskQuery, id).
		Scan(&task.ID, &task.Title, &task.Description, &task.Status, &task.Created_at, &task.Updated_at)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get task")
	}
	return &task, nil
}
func (r *repository) CreateUser(ctx context.Context, username, passwordHash string) error {
	_, err := r.pool.Exec(
		ctx,
		InsertQuery,
		username,
		passwordHash,
	)

	if err != nil {
		return fmt.Errorf("db user insert failed: %w", err)
	}

	return nil
}

func (r *repository) GetUser(ctx context.Context, username string) (*User, error) {
	var user User
	err := r.pool.QueryRow(ctx, SelectQuery, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
	)

	if err != nil {
		return nil, fmt.Errorf("db query failed: %w", err)
	}
	return &user, nil
}

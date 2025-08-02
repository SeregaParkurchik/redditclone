package pg

import (
	"context"
	"log/slog"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}
type DBOption func(cfg *pgxpool.Config)

// WithMaxConns устанавливает максимальное количество открытых подключений.
func WithMaxConns(n int32) DBOption {
	return func(cfg *pgxpool.Config) {
		cfg.MaxConns = n
	}
}

// WithMinConns устанавливает минимальное количество открытых подключений.
func WithMinConns(n int32) DBOption {
	return func(cfg *pgxpool.Config) {
		cfg.MinConns = n
	}
}

// WithConnMaxLifetime устанавливает максимальное время жизни подключения.
func WithConnMaxLifetime(d time.Duration) DBOption {
	return func(cfg *pgxpool.Config) {
		cfg.MaxConnLifetime = d
	}
}

// WithConnMaxIdleTime устанавливает максимальное время простоя подключения.
func WithConnMaxIdleTime(d time.Duration) DBOption {
	return func(cfg *pgxpool.Config) {
		cfg.MaxConnIdleTime = d
	}
}

func NewDB(ctx context.Context, connString string, logger *slog.Logger, opts ...DBOption) (*DB, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		logger.Error("Не удалось разобрать строку подключения", "ошибка", err)
		return nil, err
	}

	for _, opt := range opts {
		opt(config)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		logger.Error("Не удалось создать пул подключений к базе данных", "ошибка", err)
		return nil, err
	}

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		logger.Error("Не удалось выполнить ping базы данных", "ошибка", err)
		return nil, err
	}

	logger.Info("Пул подключений к базе данных успешно инициализирован и проверен")
	return &DB{pool: pool, logger: logger}, nil
}

type Querier interface {
	QueryOne(ctx context.Context, dest any, query string, args ...any) error
	QueryMany(ctx context.Context, dest any, query string, args ...any) error
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
	Close()
}

type Tx interface {
	QueryOne(ctx context.Context, dest any, query string, args ...any) error
	QueryMany(ctx context.Context, dest any, query string, args ...any) error
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
}

type txWrapper struct {
	tx     pgx.Tx
	logger *slog.Logger
}

func (db *DB) WithTx(ctx context.Context, fn func(tx Tx) error) (err error) {
	start := time.Now()
	pgxTx, err := db.pool.Begin(ctx)
	if err != nil {
		db.logger.Error("Не удалось начать транзакцию", "ошибка", err, "длительность", time.Since(start))
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			db.logger.Error("Паника во время транзакции, откат", "паника", p)
			if rollbackErr := pgxTx.Rollback(ctx); rollbackErr != nil {
				db.logger.Error("Не удалось откатить транзакцию после паники", "ошибка", rollbackErr)
			}
			panic(p)
		} else if err != nil {
			db.logger.Error("Ошибка во время транзакции, откат", "ошибка", err)
			if rollbackErr := pgxTx.Rollback(ctx); rollbackErr != nil {
				db.logger.Error("Не удалось откатить транзакцию после ошибки", "ошибка", rollbackErr)
			}
		} else {
			db.logger.Info("Транзакция успешно завершена, коммит")
			if commitErr := pgxTx.Commit(ctx); commitErr != nil {
				db.logger.Error("Не удалось закоммитить транзакцию", "ошибка", commitErr)
				err = commitErr
			}
		}
	}()

	tx := &txWrapper{tx: pgxTx, logger: db.logger}
	err = fn(tx)

	return err
}

func (t *txWrapper) QueryOne(ctx context.Context, dest any, query string, args ...any) error {
	start := time.Now()
	err := pgxscan.Get(ctx, t.tx, dest, query, args...)
	logAndMetricQuery(t.logger, "QueryOne", query, args, start, err)
	return err
}

func (t *txWrapper) QueryMany(ctx context.Context, dest any, query string, args ...any) error {
	start := time.Now()
	err := pgxscan.Select(ctx, t.tx, dest, query, args...)
	logAndMetricQuery(t.logger, "QueryMany", query, args, start, err)
	return err
}

func (t *txWrapper) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	start := time.Now()
	cmdTag, err := t.tx.Exec(ctx, query, args...)
	logAndMetricQuery(t.logger, "Exec", query, args, start, err)
	return cmdTag, err
}

func (db *DB) Close() {
	db.logger.Info("Закрытие пула подключений к базе данных")
	db.pool.Close()
}

func (db *DB) QueryOne(ctx context.Context, dest any, query string, args ...any) error {
	start := time.Now()
	err := pgxscan.Get(ctx, db.pool, dest, query, args...)
	logAndMetricQuery(db.logger, "QueryOne", query, args, start, err)
	return err
}

func (db *DB) QueryMany(ctx context.Context, dest any, query string, args ...any) error {
	start := time.Now()
	err := pgxscan.Select(ctx, db.pool, dest, query, args...)
	logAndMetricQuery(db.logger, "QueryMany", query, args, start, err)
	return err
}

func (db *DB) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	start := time.Now()
	cmdTag, err := db.pool.Exec(ctx, query, args...)
	logAndMetricQuery(db.logger, "Exec", query, args, start, err)
	return cmdTag, err
}

func logAndMetricQuery(logger *slog.Logger, method, query string, args []any, start time.Time, err error) {
	duration := time.Since(start)

	status := "success"
	if err != nil {
		status = "error"
		logger.Error("Сбой выполнения запроса к базе данных",
			"метод", method,
			"запрос", query,
			"аргументы", args,
			"длительность", duration,
			"ошибка", err,
		)
	} else {
		logger.Info("Запрос к базе данных выполнен успешно",
			"метод", method,
			"запрос", query,
			"аргументы", args,
			"длительность", duration,
		)
	}

	dbQueriesTotal.WithLabelValues(method, status).Inc()
	dbQueryDuration.WithLabelValues(method, status).Observe(duration.Seconds())
}

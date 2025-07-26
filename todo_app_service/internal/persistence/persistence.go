package persistence

import (
	log "Todo-List/internProject/todo_app_service/pkg/configuration"
	"context"
	"database/sql"
	"errors"

	_ "github.com/lib/pq"

	"github.com/jmoiron/sqlx"
)

type persitentCtx string

const PersistenceCtxKey persitentCtx = "persistentCtxKey"

const RetryCount int = 50

func SaveToContext(ctx context.Context, persistOp persistenceOp) context.Context {
	return context.WithValue(ctx, PersistenceCtxKey, persistOp)
}

func FromCtx(ctx context.Context) (persistenceOp, error) {
	dbCtx := ctx.Value(PersistenceCtxKey)

	if database, ok := dbCtx.(persistenceOp); ok {
		return database, nil
	}

	return nil, errors.New("unable to fetch database from context")
}

type Transactioner interface {
	BeginContext(ctx context.Context) (persistenceTx, error)
	RollbackUnlessCommitted(ctx context.Context, tx persistenceTx) bool
	PingContext(ctx context.Context) error
	Stats() sql.DBStats
}

type db struct {
	sqlDB *sqlx.DB
}

func NewSqlDb(sqlDb *sqlx.DB) *db {
	return &db{
		sqlDB: sqlDb,
	}
}

func (db *db) PingContext(ctx context.Context) error {
	return db.sqlDB.PingContext(ctx)
}

func (db *db) Stats() sql.DBStats {
	return db.sqlDB.Stats()
}

func (db *db) BeginContext(ctx context.Context) (persistenceTx, error) {
	tx, err := db.sqlDB.BeginTxx(ctx, nil)
	customTx := &Transaction{
		Tx:        tx,
		committed: false,
	}
	return persistenceTx(customTx), err
}

func (db *db) RollbackUnlessCommitted(ctx context.Context, tx persistenceTx) bool {
	customTx, ok := tx.(*Transaction)
	if !ok {
		log.C(ctx).Warn("State aware transaction is not in use")
		db.rollback(ctx, tx)
		return true
	}
	if customTx.committed {
		return false
	}
	db.rollback(ctx, customTx)
	return true
}

func (db *db) rollback(ctx context.Context, tx persistenceTx) {
	if err := tx.Rollback(); err == nil {
		log.C(ctx).Warn("Transaction rolled back")
	} else if errors.Is(err, sql.ErrTxDone) {
		log.C(ctx).Warn(err)
	}
}

type Transaction struct {
	*sqlx.Tx
	committed bool
}

//go:generate mockery --name=PersistenceTx --output=automock --outpkg=automock --case=underscore --disable-version-string --with-expecter
type persistenceTx interface {
	Commit() error
	Rollback() error
	persistenceOp
}

//go:generate mockery --name=PersistenceOp --output=automock --outpkg=automock --case=underscore --disable-version-string
type persistenceOp interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

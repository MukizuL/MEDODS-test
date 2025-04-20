package pgstorage

import (
	"MEDODS-test/internal/errs"
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"strings"
)

type PGStorage struct {
	conn   *pgxpool.Pool
	logger *zap.Logger
}

func New(ctx context.Context, dsn string, logger *zap.Logger) (*PGStorage, error) {
	dbpool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	err = dbpool.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return &PGStorage{conn: dbpool, logger: logger}, nil
}

func (P *PGStorage) SaveRefreshToken(ctx context.Context, refresh, session string) error {
	strs := strings.Split(refresh, ".")
	data, err := bcrypt.GenerateFromPassword([]byte(strs[2]), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = P.conn.Exec(ctx, "INSERT INTO tokens (session, refresh_hash) VALUES ($1, $2)", session, data)
	if err != nil {
		return err
	}

	return nil
}

func (P *PGStorage) CheckoutRefreshToken(ctx context.Context, refresh, session string) error {
	tx, err := P.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return errs.ErrInternalServerError
	}
	defer tx.Rollback(ctx)

	var refreshHash string
	err = tx.QueryRow(ctx, "SELECT refresh_hash FROM tokens WHERE session = $1", session).Scan(&refreshHash)
	if err != nil {
		return err
	}

	strs := strings.Split(refresh, ".")
	err = bcrypt.CompareHashAndPassword([]byte(refreshHash), []byte(strs[2]))
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, "DELETE FROM tokens WHERE refresh_hash = $1", refreshHash)
	if err != nil {
		return err
	}

	tx.Commit(ctx)

	return nil
}

func (P *PGStorage) Close() {
	P.conn.Close()
}

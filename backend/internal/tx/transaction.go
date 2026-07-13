package tx

import (
	"context"

	"gorm.io/gorm"
)

type txKey struct{}
type afterCommitKey struct{}

func WithDB(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey{}, db)
}

func DB(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	if db, ok := ctx.Value(txKey{}).(*gorm.DB); ok && db != nil {
		return db
	}
	return fallback
}

func AfterCommit(ctx context.Context, fn func()) {
	if fn == nil {
		return
	}
	if hooks, ok := ctx.Value(afterCommitKey{}).(*[]func()); ok && hooks != nil {
		*hooks = append(*hooks, fn)
		return
	}
	fn()
}

func WithTransaction(ctx context.Context, db *gorm.DB, fn func(ctx context.Context) error) error {
	var hooks []func()
	txCtx := context.WithValue(ctx, afterCommitKey{}, &hooks)

	err := db.WithContext(txCtx).Transaction(func(tx *gorm.DB) error {
		return fn(WithDB(txCtx, tx))
	})
	if err != nil {
		return err
	}
	for _, hook := range hooks {
		hook()
	}
	return nil
}

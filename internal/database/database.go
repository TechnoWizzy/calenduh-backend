package database

import (
	"calenduh-backend/internal/sqlc"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"log"

	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionFunc func(queries *sqlc.Queries) error

type Database struct {
	Conn    *sql.DB
	Queries *sqlc.Queries
	Pool    *pgxpool.Pool
}

var Db *Database

func Migrate(db *sql.DB) {
	// Initialize postgres migration driver
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal("could not create migration driver:", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://postgres/migrations/",
		"postgres", driver)
	if err != nil {
		log.Fatal("migration setup failed:", err)
	}

	// Apply migrations
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal("migration failed:", err)
	}
}

func New(connectionString string) error {
	var db *sql.DB
	var err error
	iterations := 0

	for ; iterations < 30; iterations++ {
		db, err = sql.Open("postgres", connectionString)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}

		err = db.Ping()
		if err != nil {
			log.Println("Unable to ping database:", err)
			time.Sleep(time.Second)
			continue
		}

		break
	}

	if iterations > 0 {
		fmt.Printf("%d iterations to connect to postgres\n", iterations+1)
	}

	pool, err := pgxpool.New(context.Background(), connectionString)
	if err != nil {
		return fmt.Errorf("failed to connect to database with pgx: %w", err)
	}

	queries := sqlc.New(pool)

	// Migrate DB
	Migrate(db)

	Db = &Database{db, queries, pool}
	return nil
}

func Transaction(ctx context.Context, next TransactionFunc) error {
	transaction, err := Db.Pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func(transaction pgx.Tx, ctx context.Context) {}(transaction, ctx)
	queries := Db.Queries.WithTx(transaction)

	if err := next(queries); err != nil {
		log.Println("could not execute transaction:", err)
		return nil
	} else {
		log.Println("executed transaction")
		return transaction.Commit(ctx)
	}
}

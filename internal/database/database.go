package database

import (
	"calenduh-backend/internal/sqlc"
	"context"

	"database/sql"
	"errors"
	"fmt"
	"log"
	// "github.com/lib/pq"

	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *sql.DB

func ConnectDB(connectionString string) {
	var err error
	DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	err = DB.Ping()
	if err != nil {
		log.Fatal("database unreachable:", err)
	}
	fmt.Println("connected successfully to postgresql")
}

type Database struct {
	Conn    *sql.DB
	Queries *sqlc.Queries
	Pool    *pgxpool.Pool
}

var Queries *sqlc.Queries

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

func New(connectionString string) (*Database, error) {
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
		return nil, fmt.Errorf("failed to connect to database with pgx: %w", err)
	}

	Queries = sqlc.New(pool)

	// Migrate DB
	Migrate(db)

	return &Database{
		db, Queries, pool,
	}, nil
}

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"mini-workflow/db/migrations"

	migrate "github.com/rubenv/sql-migrate"

	_ "github.com/lib/pq"
)

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	direction := migrate.Up
	if len(os.Args) > 1 && os.Args[1] == "down" {
		direction = migrate.Down
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getenv("DB_HOST", "localhost"),
		getenv("DB_PORT", "5432"),
		getenv("DB_USER", "user"),
		getenv("DB_PASSWORD", "password"),
		getenv("DB_NAME", "minidb"),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	src := migrate.EmbedFileSystemMigrationSource{
		FileSystem: migrations.FS,
		Root:       ".",
	}

	n, err := migrate.Exec(db, "postgres", src, direction)
	if err != nil {
		log.Fatalf("migrate: %v", err)
	}

	action := "Applied"
	if direction == migrate.Down {
		action = "Rolled back"
	}
	fmt.Printf("%s %d migration(s)\n", action, n)
}

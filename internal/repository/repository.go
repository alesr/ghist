package repository

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alesr/ghist/internal/service"
	_ "github.com/mattn/go-sqlite3"
)

const (
	dbDirname  = ".ghist"
	dbFilename = "ghist.db"

	createTableQuery = `
	CREATE TABLE IF NOT EXISTS repositories (
		name TEXT PRIMARY KEY,
		stars INTEGER,
		forks INTEGER,
		last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	upsertQuery = `
        INSERT INTO repositories (name, stars, forks, last_updated)
        VALUES (?, ?, ?, CURRENT_TIMESTAMP)
        ON CONFLICT(name) DO UPDATE SET
            stars = excluded.stars,
            forks = excluded.forks,
            last_updated = CURRENT_TIMESTAMP
    `
)

type SQLite struct{ db *sql.DB }

func NewSQLite() (*SQLite, error) {
	dbFilepath, err := getDBFilepath()
	if err != nil {
		return nil, fmt.Errorf("could not get sqlite3 db filepath: %w", err)
	}

	db, err := initDB(dbFilepath)
	if err != nil {
		return nil, fmt.Errorf("could not initialize sqlite3 db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &SQLite{db: db}, nil
}

func getDBFilepath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not get user home directory: %w", err)
	}

	ghistDir := filepath.Join(homeDir, dbDirname)
	dbPath := filepath.Join(ghistDir, dbFilename)

	// check if the database file already exists
	if _, err = os.Stat(dbPath); err == nil {
		return dbPath, nil
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("could not stat db file: %w", err)
	}

	if err := os.MkdirAll(ghistDir, 0700); err != nil {
		return "", fmt.Errorf("could not create db directory: %w", err)
	}
	return dbPath, nil
}

func initDB(dbFile string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, err
	}
	if _, err = db.Exec(createTableQuery); err != nil {
		return nil, fmt.Errorf("could not execute table query: %w", err)
	}
	return db, nil
}

func (s *SQLite) GetRepositories() ([]service.GithubRepo, error) {
	rows, err := s.db.Query("SELECT name, stars, forks FROM repositories")
	if err != nil {
		return nil, fmt.Errorf("could not query repositories: %w", err)
	}
	defer rows.Close()

	var repos []service.GithubRepo
	for rows.Next() {
		var repo service.GithubRepo
		if err := rows.Scan(&repo.Name, &repo.Stars, &repo.Forks); err != nil {
			return nil, fmt.Errorf("could not scan repository: %w", err)
		}
		repos = append(repos, repo)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during iteration: %w", err)
	}
	return repos, nil
}

func (s *SQLite) UpsertRepositories(ghRepos []service.GithubRepo) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(upsertQuery)
	if err != nil {
		return fmt.Errorf("could not prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, ghRepo := range ghRepos {
		repo := service.GithubRepo{
			Name:  ghRepo.Name,
			Stars: ghRepo.Stars,
			Forks: ghRepo.Forks,
		}
		if _, err := stmt.Exec(repo.Name, repo.Stars, repo.Forks); err != nil {
			return fmt.Errorf("could not upsert repository %s: %w", repo.Name, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %w", err)
	}
	return nil
}

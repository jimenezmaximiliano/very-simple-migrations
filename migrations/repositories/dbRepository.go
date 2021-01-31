package repositories

import (
	"database/sql"
	"fmt"
)

// DbRepository is an interface to run migrations from sql
type DbRepository interface {
	CreateMigrationsTableIfNeeded() error
	GetAlreadyRunMigrationFilePaths(migrationsDirectoryAbsolutePath string) ([]string, error)
	RunMigrationQuery(query string) error
	RegisterRunMigration(migrationFileName string) error
}

type dbRepository struct {
	db *sql.DB
}

// NewDbRepository returns an implementation of DbRepository
func NewDbRepository(db *sql.DB) DbRepository {
	return dbRepository{
		db: db,
	}
}

func (repository dbRepository) CreateMigrationsTableIfNeeded() error {
	query := `
		CREATE TABLE IF NOT EXISTS migrations (
			id INTEGER PRIMARY KEY AUTO_INCREMENT,
			migration TEXT
		);`

	_, err := repository.db.Exec(query)

	if err != nil {
		return fmt.Errorf("migrations.CreateMigrationsTable \n%w", err)
	}

	return nil
}

func (repository dbRepository) GetAlreadyRunMigrationFilePaths(migrationsDirectoryAbsolutePath string) ([]string, error) {
	rows, err := repository.db.Query("SELECT migration FROM migrations")

	if err != nil {
		return nil, fmt.Errorf("migrations.getMigrationsFromTheMigrationsTable \n%w", err)
	}

	defer rows.Close()

	return getMigrationPathsFromRows(rows, migrationsDirectoryAbsolutePath)
}

func getMigrationPathsFromRows(rows *sql.Rows, migrationsDirectoryAbsolutePath string) ([]string, error) {
	var migrationsAlreadyRun []string

	for rows.Next() {
		migrationFileName := ""
		err := rows.Scan(&migrationFileName)

		if err != nil {
			return migrationsAlreadyRun, fmt.Errorf("migrations.readMigrationRowFromMigrationsTable \n%w", err)
		}

		currentMigrationAbsolutePath := migrationsDirectoryAbsolutePath + migrationFileName
		migrationsAlreadyRun = append(migrationsAlreadyRun, currentMigrationAbsolutePath)
	}

	return migrationsAlreadyRun, nil
}

func (repository dbRepository) RunMigrationQuery(query string) error {
	_, err := repository.db.Exec(query)

	return err
}

func (repository dbRepository) RegisterRunMigration(migrationFileName string) error {
	_, err := repository.db.Exec("INSERT INTO migrations (migration) VALUES (?)", migrationFileName)

	return err
}
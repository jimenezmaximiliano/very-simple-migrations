package very_simple_migrations

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type MigrationResult struct {
	SuccessfulMigrations []string
	FailedMigrations []string
}

func Run(db *sql.DB, migrationsAbsolutePath string) (MigrationResult, error) {
	result := MigrationResult{
		SuccessfulMigrations: make([]string, 0),
		FailedMigrations: make([]string, 0),
	}

	err := createMigrationsTableIfNeeded(db)

	if err != nil {
		return result, err
	}

	migrationFilePaths, err := getMigrationFilePaths(migrationsAbsolutePath)

	if err != nil {
		return result, err
	}

	alreadyRunMigrationFilePaths, err := getAlreadyRunMigrationFilePaths(db, migrationsAbsolutePath)

	if err != nil {
		return result, err
	}

	migrationFilePathsToRun := getMigrationFilePathsToRun(migrationFilePaths, alreadyRunMigrationFilePaths)

	if len(migrationFilePathsToRun) == 0  {
		return result, nil
	}

	return runMigrations(db, migrationFilePathsToRun, result)
}

func createMigrationsTableIfNeeded(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS migrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			migration TEXT
		);`

	_, err := db.Exec(query)

	if err != nil {
		return fmt.Errorf("verySimpleMigrations.CreateMigrationsTable \n%w", err)
	}

	return nil
}

func getMigrationFilePaths(migrationsAbsolutePath string) ([]string, error) {
	if migrationsAbsolutePath[len(migrationsAbsolutePath)-1:] != "/" {
		migrationsAbsolutePath += "/"
	}

	var migrationPaths []string

	files, err := ioutil.ReadDir(migrationsAbsolutePath)

	if err != nil {
		return migrationPaths, fmt.Errorf("verySimpleMigrations.readMigrationsPath (path: %s) \n%w", migrationsAbsolutePath, err)
	}

	for _, file := range files {
		if isNotASqlFile(file) {
			continue
		}

		currentMigrationPath := migrationsAbsolutePath + file.Name()
		migrationPaths = append(migrationPaths, currentMigrationPath)
	}

	return migrationPaths, nil
}

func isNotASqlFile(file os.FileInfo) bool {
	fileName := file.Name()
	fileNameLength := len(fileName)

	return file.IsDir() || fileNameLength <= 4 || fileName[(fileNameLength - 4):] != ".sql"
}

func getAlreadyRunMigrationFilePaths(db *sql.DB, migrationsAbsolutePath string) ([]string, error) {
	var migrationsAlreadyRun []string

	rows, err := db.Query("SELECT migration FROM migrations")

	if err != nil {
		return migrationsAlreadyRun, fmt.Errorf("verySimpleMigrations.getMigrationsFromTheMigrationsTable \n%w", err)
	}

	defer rows.Close()

	for rows.Next() {
		migrationFileName := ""
		err := rows.Scan(&migrationFileName)

		if err != nil {
			return migrationsAlreadyRun, fmt.Errorf("verySimpleMigrations.readMigrationRowFromMigrationsTable \n%w", err)
		}

		currentMigrationAbsolutePath := migrationsAbsolutePath + migrationFileName
		migrationsAlreadyRun = append(migrationsAlreadyRun, currentMigrationAbsolutePath)
	}

	return migrationsAlreadyRun, nil
}

func getMigrationFilePathsToRun(all []string, alreadyRun []string) []string {
	var migrationsToRun []string

	for _, migration := range all {
		if inSlice(alreadyRun, migration) {
			continue
		}

		migrationsToRun = append(migrationsToRun, migration)
	}

	return migrationsToRun
}

func inSlice(slice []string, element string) bool {
	for _, currentElement := range slice {
		if currentElement == element {
			return true
		}
	}

	return false
}

func runMigrations(db *sql.DB, migrations []string, result MigrationResult) (MigrationResult, error) {
	for _, migration := range migrations {
		query, err := ioutil.ReadFile(migration)

		if err != nil {
			result.FailedMigrations = append(result.FailedMigrations, migration)
			return result, fmt.Errorf("verySimpleMigrations.readMigrationFile (%s) \n%w", migration, err)
		}

		_, err = db.Exec(string(query))

		if err != nil {
			result.FailedMigrations = append(result.FailedMigrations, migration)
			return result, fmt.Errorf("verySimpleMigrations.runMigration (%s) \n%w", migration, err)
		}

		migrationFileNameParts := strings.Split(migration, "/")
		migrationFileName := migrationFileNameParts[len(migrationFileNameParts) - 1]

		_, err = db.Exec("INSERT INTO migrations (migration) VALUES (?)", migrationFileName)

		if err != nil {
			return result, fmt.Errorf("verySimpleMigrations.insertMigrationIntoMigrationsTableAfterRunningIt (%s) \n%w", migration, err)
		}

		result.SuccessfulMigrations = append(result.SuccessfulMigrations, migration)
	}

	return result, nil
}
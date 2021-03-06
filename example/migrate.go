package main

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	"github.com/jimenezmaximiliano/migrations"
)

func main() {
	migrations.RunMigrationsCommand(func() (*sql.DB, error) {
		return sql.Open("mysql", "user:password@/db?multiStatements=true")
	})
}

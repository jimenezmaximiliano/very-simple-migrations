package migration

import (
	"fmt"
	"sort"
	"strings"
)

const (
	StatusUnknown    = 0
	StatusNotRun     = 1
	StatusSuccessful = 2
	StatusFailed     = -1
)

type Migration interface {
	GetAbsolutePath() string
	GetName() string
	GetStatus() int8
	ShouldBeRun() bool
	GetQuery() string
	NewAsFailed() Migration
	NewAsSuccessful() Migration
	WasSuccessful() bool
	HasFailed() bool
	ShouldBeRunFirst(anotherMigration Migration) bool
}

type migration struct {
	absolutePath string
	name         string
	status       int8
	query        string
}

func (migration migration) GetAbsolutePath() string {
	return migration.absolutePath
}

func (migration migration) GetName() string {
	return migration.name
}

func (migration migration) GetStatus() int8 {
	return migration.status
}

func (migration migration) ShouldBeRun() bool {
	return migration.status != StatusSuccessful
}

func (migration migration) WasSuccessful() bool {
	return migration.status == StatusSuccessful
}

func (migration migration) HasFailed() bool {
	return migration.status == StatusFailed
}

func (migration migration) GetQuery() string {
	return migration.query
}

func (migration migration) NewAsFailed() Migration {
	newMigration, _ := NewMigration(migration.GetAbsolutePath(), migration.GetQuery(), StatusFailed)

	return newMigration
}

func (migration migration) NewAsSuccessful() Migration {
	newMigration, _ := NewMigration(migration.GetAbsolutePath(), migration.GetQuery(), StatusSuccessful)

	return newMigration
}

func (migration migration) ShouldBeRunFirst(anotherMigration Migration) bool {
	names := []string{
		migration.name,
		anotherMigration.GetName(),
	}

	sort.Strings(names)

	return names[0] == migration.name
}

func NewMigration(absolutePath string, query string, status int8) (Migration, error) {

	if status < -1 || status > 2 {
		return migration{}, fmt.Errorf("migrations.migration.New.invalidStatus (status: %d)", status)
	}

	return migration{
		absolutePath: absolutePath,
		name:         extractFileName(absolutePath),
		status:       status,
		query:        query,
	}, nil
}

func extractFileName(absolutePath string) string {
	absolutePathParts := strings.Split(absolutePath, "/")

	return getSliceLastElement(absolutePathParts)
}

func getSliceLastElement(theSlice []string) string {
	lastIndex := len(theSlice) - 1

	return theSlice[lastIndex]
}
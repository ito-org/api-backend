package main

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// NewDBConnection creates and tests a new db connection and returns it.
func NewDBConnection(dbHost, dbUser, dbPassword, dbName string) (*DBConnection, error) {
	connStr := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s sslmode=disabled",
		dbHost,
		dbUser,
		dbPassword,
		dbName,
	)

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		fmt.Printf("Failed to connect to Postgres database: %s\n", err.Error())
		return nil, err
	}
	return &DBConnection{db}, err
}

// DBConnection implements several functions for fetching and manipulation
// of reports in the database.
type DBConnection struct {
	*sqlx.DB
}

func (db *DBConnection) insertReport(report *Report) error {
	// TODO
	return nil
}

func (db *DBConnection) getReports() ([]*Report, error) {
	// TODO
	return nil, nil
}

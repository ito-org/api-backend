package main

import (
	"fmt"

	"github.com/ito-org/go-backend/tcn"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// NewDBConnection creates and tests a new db connection and returns it.
func NewDBConnection(dbHost, dbUser, dbPassword, dbName string) (*DBConnection, error) {
	connStr := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s sslmode=disable",
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

func (db *DBConnection) insertMemo(memo *tcn.Memo) (uint64, error) {
	var newID uint64
	if err := db.QueryRowx(
		`
		INSERT INTO
		Memo(mtype, mlen, mdata)
		VALUES($1, $2, $3)
		RETURNING id;
		`,
		memo.Type,
		memo.Len,
		memo.Data[:],
	).Scan(&newID); err != nil {
		fmt.Printf("Failed to insert memo into database: %s\n", err.Error())
		return 0, err
	}
	return newID, nil
}

func (db *DBConnection) insertReport(report *tcn.Report) error {
	memoID, err := db.insertMemo(report.Memo)
	if err != nil {
		return err
	}

	_, err = db.Exec(
		`
	INSERT INTO
	Report(rvk, tck_bytes, j_1, j_2, memo_id)
	VALUES($1, $2, $3, $4, $5);
	`,
		report.RVK,
		report.TCKBytes[:],
		report.J1,
		report.J2,
		memoID,
	)
	if err != nil {
		fmt.Printf("Failed to insert report into database: %s\n", err.Error())
		return err
	}
	return nil
}

func (db *DBConnection) getReports() ([]*tcn.Report, error) {
	reports := []*tcn.Report{}

	rows, err := db.Queryx(
		`SELECT r.rvk, r.tck_bytes, r.j_1, r.j_2, m.mtype, m.mlen, m.mdata
		FROM Report r
		JOIN Memo m ON r.memo_id = m.id;
		`,
	)
	if err != nil {
		fmt.Printf("Failed to get reports from database: %s\n", err.Error())
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		report := &tcn.Report{
			TCKBytes: [32]uint8{},
			Memo:     &tcn.Memo{},
		}
		tckBytesDest := []byte{}
		if err := rows.Scan(
			&report.RVK,
			&tckBytesDest,
			&report.J1,
			&report.J2,
			&report.Memo.Type,
			&report.Memo.Len,
			&report.Memo.Data,
		); err != nil {
			fmt.Printf("Failed to scan report: %s\n", err.Error())
			return nil, err
		}

		copy(report.TCKBytes[:], tckBytesDest[:32])
		reports = append(reports, report)
	}

	return reports, nil
}

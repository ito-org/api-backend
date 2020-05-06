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

func (db *DBConnection) insertReport(report *tcn.Report) (uint64, error) {
	memoID, err := db.insertMemo(report.Memo)
	if err != nil {
		return 0, err
	}

	var newID uint64

	if err = db.QueryRowx(
		`
	INSERT INTO
	Report(rvk, tck_bytes, j_1, j_2, memo_id)
	VALUES($1, $2, $3, $4, $5)
	RETURNING id;
	`,
		report.RVK,
		report.TCKBytes[:],
		report.J1,
		report.J2,
		memoID,
	).Scan(&newID); err != nil {
		fmt.Printf("Failed to insert report into database: %s\n", err.Error())
		return 0, err
	}
	return newID, nil
}

func (db *DBConnection) insertSignedReport(signedReport *tcn.SignedReport) error {
	reportID, err := db.insertReport(signedReport.Report)
	if err != nil {
		return err
	}

	_, err = db.Exec(
		`
		INSERT INTO
		SignedReport(report_id, sig)
		VALUES($1, $2)
		`,
		reportID,
		signedReport.Sig[:],
	)
	if err != nil {
		fmt.Printf("Failed to insert signed report into database: %s\n", err.Error())
		return err
	}
	return nil
}

func (db *DBConnection) scanSignedReports(rows *sqlx.Rows) ([]*tcn.SignedReport, error) {
	signedReports := []*tcn.SignedReport{}
	for rows.Next() {
		signedReport := &tcn.SignedReport{
			Report: &tcn.Report{
				TCKBytes: [32]uint8{},
				Memo:     &tcn.Memo{},
			},
			Sig: []byte{},
		}
		tckBytesDest := []byte{}
		if err := rows.Scan(
			&signedReport.Report.RVK,
			&tckBytesDest,
			&signedReport.Report.J1,
			&signedReport.Report.J2,
			&signedReport.Report.Memo.Type,
			&signedReport.Report.Memo.Len,
			&signedReport.Report.Memo.Data,
			&signedReport.Sig,
		); err != nil {
			fmt.Printf("Failed to scan signed report: %s\n", err.Error())
			return nil, err
		}

		copy(signedReport.Report.TCKBytes[:], tckBytesDest[:32])
		signedReports = append(signedReports, signedReport)
	}
	return signedReports, nil
}

func (db *DBConnection) getSignedReports() ([]*tcn.SignedReport, error) {
	rows, err := db.Queryx(
		`
		SELECT r.rvk, r.tck_bytes, r.j_1, r.j_2, m.mtype, m.mlen, m.mdata, sr.sig
		FROM SignedReport sr
		JOIN Report r ON sr.report_id = r.id
		JOIN Memo m ON r.memo_id = m.id;
		`,
	)
	if err != nil {
		fmt.Printf("Failed to get signed reports from database: %s\n", err.Error())
		return nil, err
	}
	defer rows.Close()
	signedReports, err := db.scanSignedReports(rows)
	if err != nil {
		return nil, err
	}
	return signedReports, nil
}

// getNewSignedReports returns all signed reports that were made after lastReport.
func (db *DBConnection) getNewSignedReports(lastReport *tcn.Report) ([]*tcn.SignedReport, error) {
	rows, err := db.Queryx(
		`
		SELECT r.rvk, r.tck_bytes, r.j_1, r.j_2, m.mtype, m.mlen, m.mdata, sr.sig
		FROM SignedReport sr
		JOIN Report r ON sr.report_id = r.id
		WHERE r.timestamp > (
			SELECT r2.timestamp
			FROM Report r2
			WHERE r2.rvk = $1
			AND r2.tck_bytes = $2
			AND r2.j_1 = $3
			AND r2.j_1 = $4
		);
		`,
		lastReport.RVK,
		lastReport.TCKBytes[:],
		lastReport.J1,
		lastReport.J2,
	)
	if err != nil {
		fmt.Printf("Failed to get signed reports from database: %s\n", err.Error())
		return nil, err
	}
	defer rows.Close()
	signedReports, err := db.scanSignedReports(rows)
	if err != nil {
		return nil, err
	}
	return signedReports, nil
}

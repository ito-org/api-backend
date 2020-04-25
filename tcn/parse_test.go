package tcn_test

import (
	"testing"

	"github.com/ito-org/go-backend/tcn"
	"github.com/stretchr/testify/assert"
)

func TestGetReports(t *testing.T) {
	reports := [5]*tcn.Report{}
	for i := 0; i < 5; i++ {
		_, _, report, _ := tcn.GenerateReport(0, 1, []byte("symptom data"))
		reports[i] = report
	}

	reportBytes := []byte{}
	for _, r := range reports {
		b, err := r.Bytes()
		if err != nil {
			t.Error(err.Error())
			return
		}
		reportBytes = append(reportBytes, b...)
	}

	retReports := tcn.GetReports(reportBytes)

	assert.Len(t, retReports, len(reports))
	for i, rr := range retReports {
		assert.EqualValues(t, reports[i], rr)
	}
}

func TestGetReport(t *testing.T) {
	_, _, report, err := tcn.GenerateReport(0, 1, []byte("symptom data"))
	if err != nil {
		t.Error(err.Error())
		return
	}

	rb, err := report.Bytes()
	if err != nil {
		t.Error(err.Error())
		return
	}

	retReport := tcn.GetReport(rb)

	assert.EqualValues(t, report, retReport)
}

func TestGetSignedReport(t *testing.T) {
	_, rak, report, err := tcn.GenerateReport(0, 4, []byte("sympton data"))
	if err != nil {
		t.Error(err.Error())
		return
	}
	signedReport, err := tcn.GenerateSignedReport(rak, report)
	if err != nil {
		t.Error(err.Error())
		return
	}

	srb, err := signedReport.Bytes()
	if err != nil {
		t.Error(err.Error())
		return
	}

	retSignedReport, err := tcn.GetSignedReport(srb)

	assert.EqualValues(t, signedReport, retSignedReport)
}

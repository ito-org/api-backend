package main

import (
	"bytes"
	"crypto"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ito-org/go-backend/tcn"
	"github.com/stretchr/testify/assert"
)

var handler *TCNReportHandler

// Init function before every test
func TestMain(m *testing.M) {
	// Initialize the database connection and the handler structure so we can
	// call the handler functions directly instead of making actual HTTP
	// requests. This allows us to create and the database connection which
	// would otherwise not happen.

	dbHost, dbName, dbUser, dbPassword := readPostgresSettings()

	dbConn, err := NewDBConnection(dbHost, dbUser, dbPassword, dbName)
	if err != nil {
		panic(err.Error())
	}

	handler = &TCNReportHandler{
		dbConn: dbConn,
	}
	code := m.Run()
	os.Exit(code)
}

func getGetRequest() (*httptest.ResponseRecorder, *http.Request) {
	return getGetRequestWithParam("")
}

func getGetRequestWithParam(from string) (*httptest.ResponseRecorder, *http.Request) {
	path := "/tcnreport"
	if from != "" {
		path += fmt.Sprintf("?from=%s", from)
	}

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", path, nil)
	return rec, req
}

func getPostRequest(data []byte) (*httptest.ResponseRecorder, *http.Request) {
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/tcnreport", bytes.NewReader(data))
	return rec, req
}

func TestPostTCNReport(t *testing.T) {
	_, rak, report, err := tcn.GenerateReport(0, 1, []byte("symptom data"))
	if err != nil {
		t.Error(err)
		return
	}

	signedReport, err := tcn.GenerateSignedReport(rak, report)
	if err != nil {
		t.Error(err)
		return
	}

	b, err := signedReport.Bytes()
	if err != nil {
		t.Error(err)
		return
	}

	rec, req := getPostRequest(b)
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = req
	handler.postTCNReport(ctx)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestPostTCNReportInvalidSig(t *testing.T) {
	// Store just the report here since we're going to sign it with a different
	// key
	_, _, report, err := tcn.GenerateReport(0, 1, nil)
	if err != nil {
		t.Error(err)
	}

	// Generate second private key to sign with so we can force an error to
	// happen
	_, rak2, _, err := tcn.GenerateReport(0, 1, nil)
	if err != nil {
		t.Error(err)
		return
	}

	rb, err := report.Bytes()
	if err != nil {
		t.Error(err)
		return
	}

	fakeSignedReport, err := tcn.GenerateSignedReport(rak2, report) // Note: wrong key used here
	if err != nil {
		t.Error(err)
		return
	}

	// Manually concatenate the byte array that's sent to the server
	b := []byte{}
	b = append(b, rb...)
	b = append(b, fakeSignedReport.Sig...)

	rec, req := getPostRequest(b)
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = req
	handler.postTCNReport(ctx)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	defer rec.Result().Body.Close()
	respData, err := ioutil.ReadAll(rec.Result().Body)
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, reportVerificationError, string(respData))
}

func TestPostTCNInvalidType(t *testing.T) {
	_, rak, report, err := tcn.GenerateReport(0, 1, nil)
	if err != nil {
		t.Error(err)
		return
	}

	// Not ito memo type
	report.Memo.Type = 0x1

	signedReport, err := tcn.GenerateSignedReport(rak, report)
	if err != nil {
		t.Error(err)
		return
	}

	b, err := signedReport.Bytes()
	if err != nil {
		t.Error(err)
		return
	}

	rec, req := getPostRequest(b)
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = req
	handler.postTCNReport(ctx)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPostTCNInvalidLength(t *testing.T) {
	_, rak, report, err := tcn.GenerateReport(0, 1, nil)
	if err != nil {
		t.Error(err)
		return
	}

	report.Memo.Len = 0
	report.Memo.Data = nil

	rb, err := report.Bytes()
	if err != nil {
		t.Error(err)
		return
	}

	rb = rb[1:] // This is where the report gets its invalid length

	sig, err := rak.Sign(nil, rb, crypto.Hash(0))
	if err != nil {
		t.Error(err)
		return
	}

	signedReport := &tcn.SignedReport{
		Report: report,
		Sig:    sig,
	}

	var b []byte
	b = append(b, rb...)
	b = append(b, signedReport.Sig...)

	rec, req := getPostRequest(b)
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = req
	handler.postTCNReport(ctx)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetTCNReports(t *testing.T) {
	signedReports := [5]*tcn.SignedReport{}
	for i := 0; i < 5; i++ {
		_, rak, report, _ := tcn.GenerateReport(0, 1, []byte("symptom data"))
		signedReport, err := tcn.GenerateSignedReport(rak, report)
		if err != nil {
			t.Error(err.Error())
			return
		}

		signedReportBytes, err := signedReport.Bytes()
		if err != nil {
			t.Error(err.Error())
			return
		}

		// POST reports
		rec, req := getPostRequest(signedReportBytes)
		ctx, _ := gin.CreateTestContext(rec)
		ctx.Request = req
		handler.postTCNReport(ctx)

		signedReports[i] = signedReport
	}

	// GET reports
	rec, req := getGetRequest()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = req
	handler.getTCNReport(ctx)
	defer rec.Result().Body.Close()

	body, err := ioutil.ReadAll(rec.Result().Body)
	if err != nil {
		t.Error(err.Error())
	}

	if len(body) == 0 {
		t.Error("Body is empty")
		return
	}

	// Retrieve the signed reports from the handler function's response
	retSignedReports := tcn.GetSignedReports(body)
	if err != nil {
		t.Error(err.Error())
		return
	}

	found := 0
	for _, r := range signedReports {
		for _, rr := range retSignedReports {
			if reflect.DeepEqual(r, rr) {
				found++
			}
		}
	}

	assert.Equal(t, len(signedReports), found)
}

func postSignedReports(signedReportBytes []byte) {
	rec, req := getPostRequest(signedReportBytes)
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = req
	handler.postTCNReport(ctx)
}

func TestGetNewTCNReports(t *testing.T) {
	signedReports := [5]*tcn.SignedReport{}
	for i := 0; i < 5; i++ {
		_, rak, report, _ := tcn.GenerateReport(0, 1, []byte("symptom data"))
		signedReport, err := tcn.GenerateSignedReport(rak, report)
		if err != nil {
			t.Error(err.Error())
			return
		}

		signedReportBytes, err := signedReport.Bytes()
		if err != nil {
			t.Error(err.Error())
			return
		}

		postSignedReports(signedReportBytes)
		if i == 2 {
			// Post it twice
			postSignedReports(signedReportBytes)
		}
		signedReports[i] = signedReport
	}

	// Prepare the hex values for reports 2 and 3
	// 2 is the one that was inserted twice and 3 was insert once.
	sr2Bytes, _ := signedReports[2].Bytes()
	hex2Dst := make([]byte, hex.EncodedLen(len(sr2Bytes)))
	hex.Encode(hex2Dst, sr2Bytes)

	// GET reports after second report (second report was inserted twice so
	// the returned reports should contain the copy of that report that was
	// inserted later, but also the rest of the reports)
	rec, req := getGetRequestWithParam(string(hex2Dst))
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = req
	handler.getTCNReport(ctx)
	defer rec.Result().Body.Close()

	body, err := ioutil.ReadAll(rec.Result().Body)
	if err != nil {
		t.Error(err.Error())
	}

	if len(body) == 0 {
		t.Error("Body is empty")
		return
	}

	// Retrieve the signed reports from the handler function's response
	retSignedReports := tcn.GetSignedReports(body)
	if err != nil {
		t.Error(err.Error())
		return
	}

	found := 0
	for _, r := range signedReports {
		for _, rr := range retSignedReports {
			if reflect.DeepEqual(r, rr) {
				found++
			}
		}
	}

	assert.Equal(t, len(signedReports[2:]), found)
}

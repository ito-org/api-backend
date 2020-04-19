package main

import (
	"bytes"
	"crypto"
	"crypto/ed25519"
	"crypto/sha256"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ito-org/go-backend/tcn"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

var handler *TCNReportHandler

// Init function before every test
func TestMain(m *testing.M) {
	// Initialize the database connection and the handler structure so we can
	// call the handler functions directly instead of making actual HTTP
	// requests. This allows us to create and the database connection which
	// would otherwise not happen.

	if err := godotenv.Load(); err != nil {
		panic(err.Error())
	}

	dbName := os.Getenv("POSTGRES_DB")
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")

	if dbName == "" || dbUser == "" || dbPassword == "" {
		panic("Error loading environment variables")
	}

	dbConn, err := NewDBConnection("localhost", dbUser, dbPassword, dbName)
	if err != nil {
		panic(err.Error())
	}

	handler = &TCNReportHandler{
		dbConn: dbConn,
	}
	code := m.Run()
	os.Exit(code)
}

func getPostRequest(data []byte) (*httptest.ResponseRecorder, *http.Request) {
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/tcnreport", bytes.NewReader(data))
	return rec, req
}

func generateMemo(content []byte) (*tcn.Memo, error) {
	if len(content) > 255 {
		return nil, errors.New("Data field contains too many bytes")
	}

	return &tcn.Memo{
		Type: tcn.ITOMemoCode,
		Len:  uint8(len(content)),
		Data: content,
	}, nil
}

func generateRakAndReport() (*ed25519.PrivateKey, *tcn.Report, error) {
	rvk, rak, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, nil, err
	}

	tck0Hash := sha256.New()
	tck0Hash.Write([]byte(tcn.HTCKDomainSep))
	tck0Hash.Write(rak)

	tck0Bytes := [32]byte{}
	copy(tck0Bytes[:32], tck0Hash.Sum(nil))

	tck0 := &tcn.TemporaryContactKey{
		Index:    0,
		RVK:      rvk,
		TCKBytes: tck0Bytes,
	}

	tck1, err := tck0.Ratchet()
	if err != nil {
		return nil, nil, err
	}

	memo, err := generateMemo([]byte("symptom data"))
	if err != nil {
		return nil, nil, err
	}

	report := &tcn.Report{
		RVK:      rvk,
		TCKBytes: tck1.TCKBytes,
		J1:       0,
		J2:       1,
		Memo:     memo,
	}

	return &rak, report, nil
}

func generateSignedReport(rak *ed25519.PrivateKey, report *tcn.Report) (*tcn.SignedReport, error) {
	b, err := report.Bytes()
	if err != nil {
		return nil, err
	}

	sig, err := rak.Sign(nil, b, crypto.Hash(0))
	if err != nil {
		return nil, err
	}

	return &tcn.SignedReport{
		Report: report,
		Sig:    sig,
	}, nil
}

func TestPostTCNReport(t *testing.T) {
	rak, report, err := generateRakAndReport()
	if err != nil {
		t.Error(err)
	}

	signedReport, err := generateSignedReport(rak, report)
	if err != nil {
		t.Error(err)
	}

	b, err := signedReport.Bytes()
	if err != nil {
		t.Error(err)
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
	_, report, err := generateRakAndReport()
	if err != nil {
		t.Error(err)
	}

	// Generate second private key to sign with so we can force an error to
	// happen
	rak2, _, err := generateRakAndReport()
	if err != nil {
		t.Error(err)
	}

	rb, err := report.Bytes()
	if err != nil {
		t.Error(err)
	}

	fakeSignedReport, err := generateSignedReport(rak2, report) // Note: wrong key used here
	if err != nil {
		t.Error(err)
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
	}
	assert.Equal(t, reportVerificationError, string(respData))
}

func TestPostTCNInvalidType(t *testing.T) {
	rak, report, err := generateRakAndReport()
	if err != nil {
		t.Error(err)
	}

	// Not ito memo type
	report.Memo.Type = 0x1

	signedReport, err := generateSignedReport(rak, report)
	if err != nil {
		t.Error(err)
	}

	b, err := signedReport.Bytes()
	if err != nil {
		t.Error(err)
	}

	rec, req := getPostRequest(b)
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = req
	handler.postTCNReport(ctx)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPostTCNInvalidLength(t *testing.T) {
	rak, report, err := generateRakAndReport()
	if err != nil {
		t.Error(err)
	}

	report.Memo.Len = 0
	report.Memo.Data = nil

	rb, err := report.Bytes()
	if err != nil {
		t.Error(err)
	}

	rb = rb[1:]

	sig, err := rak.Sign(nil, rb, crypto.Hash(0))
	if err != nil {
		t.Error(err)
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

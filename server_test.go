package main

import (
	"bytes"
	"crypto"
	"crypto/ed25519"
	"crypto/sha256"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ito-org/go-backend/tcn"
	"github.com/stretchr/testify/assert"
)

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

	router := GetRouter("8000", nil)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/tcnreport", bytes.NewReader(b))
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

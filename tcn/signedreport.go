package tcn

import (
	"crypto"
	"crypto/ed25519"
)

const (
	// SignedReportMinLength defines a signed report's minimum length in bytes
	SignedReportMinLength = 70
)

// SignedReport contains a report and the corresponding signature. The client
// sends this to the server.
type SignedReport struct {
	*Report
	// This is an ed25519 signature in byte array form
	// The ed25519 package returns a byte array as the signature
	// here: https://golang.org/pkg/crypto/ed25519/#PrivateKey.Sign
	Sig []byte `db:"sig"`
}

// Bytes converts sr to a concatenated byte array representation.
func (sr *SignedReport) Bytes() ([]byte, error) {
	var data []byte
	b, err := sr.Report.Bytes()
	if err != nil {
		return nil, err
	}
	data = append(data, b...)
	data = append(data, sr.Sig...)
	return data, nil
}

// GenerateSignedReport signs a report with rak and returns the signed report.
func GenerateSignedReport(rak *ed25519.PrivateKey, report *Report) (*SignedReport, error) {
	b, err := report.Bytes()
	if err != nil {
		return nil, err
	}

	sig, err := rak.Sign(nil, b, crypto.Hash(0))
	if err != nil {
		return nil, err
	}

	return &SignedReport{
		Report: report,
		Sig:    sig,
	}, nil
}

// Verify uses ed25519's Verify function to verify the signature over the
// report.
func (sr *SignedReport) Verify() (bool, error) {
	reportBytes, err := sr.Report.Bytes()
	if err != nil {
		return false, err
	}
	return ed25519.Verify(sr.Report.RVK, reportBytes, sr.Sig), nil
}

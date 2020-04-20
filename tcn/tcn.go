package tcn

import (
	"crypto"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"math"
)

const (
	// ITOMemoCode is the code that marks a report as an ito report in the
	// memo.
	ITOMemoCode = 0x2

	// HTCKDomainSep is the domain separator used for the domain-separated hash
	// function.
	HTCKDomainSep = "H_TCK"

	// SignedReportMinLength defines a signed report's minimum length in bytes
	SignedReportMinLength = 70
)

// Report represents a report as described in the TCN protocol:
// https://github.com/TCNCoalition/TCN#reporting
type Report struct {
	RVK      ed25519.PublicKey `db:"rvk"`
	TCKBytes [32]byte          `db:"tck_bytes"`
	J1       uint16            `db:"j_1"`
	J2       uint16            `db:"j_2"`
	*Memo
}

// Memo represents a memo data set as described in the TCN protocol:
// https://github.com/TCNCoalition/TCN#reporting
type Memo struct {
	Type uint8   `db:"mtype"`
	Len  uint8   `db:"mlen"`
	Data []uint8 `db:"mdata"`
}

// Bytes converts r to a concatenated byte array represention.
func (r *Report) Bytes() ([]byte, error) {
	var data []byte
	data = append(data, r.RVK...)
	data = append(data, r.TCKBytes[:]...)

	j1Bytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(j1Bytes, r.J1)
	j2Bytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(j2Bytes, r.J2)
	data = append(data, j1Bytes...)
	data = append(data, j2Bytes...)

	if r.Memo == nil {
		return nil, errors.New("Failed to create byte representation of report: memo field is null")
	}

	// Memo
	data = append(data, r.Memo.Type)
	data = append(data, r.Memo.Len)
	data = append(data, r.Memo.Data...)

	return data, nil
}

// GenerateMemo returns a memo instance with the given content.
func GenerateMemo(content []byte) (*Memo, error) {
	if len(content) > 255 {
		return nil, errors.New("Data field contains too many bytes")
	}

	var c []byte
	// If content is nil, we don't want the data field in the memo to be nil
	// but empty instead.
	if content != nil {
		c = content
	} else {
		c = []byte{}
	}

	return &Memo{
		Type: ITOMemoCode,
		Len:  uint8(len(content)),
		Data: c,
	}, nil
}

// GenerateReport creates a public key, private key, and report according to TCN.
func GenerateReport(j1, j2 uint16, memoData []byte) (*ed25519.PublicKey, *ed25519.PrivateKey, *Report, error) {
	rvk, rak, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, nil, nil, err
	}

	tck0Hash := sha256.New()
	tck0Hash.Write([]byte(HTCKDomainSep))
	tck0Hash.Write(rak)

	tck0Bytes := [32]byte{}
	copy(tck0Bytes[:32], tck0Hash.Sum(nil))

	tck0 := &TemporaryContactKey{
		Index:    0,
		RVK:      rvk,
		TCKBytes: tck0Bytes,
	}

	tck1, err := tck0.Ratchet()
	if err != nil {
		return nil, nil, nil, err
	}

	memo, err := GenerateMemo(memoData)
	if err != nil {
		return nil, nil, nil, err
	}

	report := &Report{
		RVK:      rvk,
		TCKBytes: tck1.TCKBytes,
		J1:       j1,
		J2:       j2,
		Memo:     memo,
	}

	return &rvk, &rak, report, nil
}

// SignedReport contains a report and the corresponding signature. The client
// sends this to the server.
type SignedReport struct {
	Report *Report
	// This is a ed25519 signature in byte array form
	// The ed25519 package returns a byte array as the signature
	// here: https://golang.org/pkg/crypto/ed25519/#PrivateKey.Sign
	Sig []byte
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

// TemporaryContactNumber is a pseudorandom 128-bit value broadcast to nearby
// devices over Bluetooth
type TemporaryContactNumber [16]uint8

// TemporaryContactKey is a ratcheting key used to derive temporary contact
// numbers.
type TemporaryContactKey struct {
	Index    uint16
	RVK      ed25519.PublicKey
	TCKBytes [32]byte
}

// Ratchet the key forward, producing a new key for a new temporary
// contact number.
func (tck *TemporaryContactKey) Ratchet() (*TemporaryContactKey, error) {
	nextHash := sha256.New()
	nextHash.Write([]byte(HTCKDomainSep))
	nextHash.Write(tck.RVK)
	nextHash.Write(tck.TCKBytes[:])

	if tck.Index == math.MaxUint16 {
		return nil, errors.New("rak should be rotated")
	}

	newTCKBytes := [32]byte{}
	copy(newTCKBytes[:32], nextHash.Sum(nil))

	return &TemporaryContactKey{
		Index:    tck.Index + 1,
		RVK:      tck.RVK,
		TCKBytes: newTCKBytes,
	}, nil
}

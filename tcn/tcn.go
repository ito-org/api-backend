package tcn

import (
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
	RVK      ed25519.PublicKey
	TCKBytes [32]byte
	J1       uint16
	J2       uint16
	Memo     *Memo
}

// Memo represents a memo data set as described in the TCN protocol:
// https://github.com/TCNCoalition/TCN#reporting
type Memo struct {
	Type uint8
	Len  uint8
	Data []uint8
}

// Bytes converts r to a concatenated byte array representions.
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

package tcn

import (
	"crypto/ed25519"
	"crypto/sha256"
	"errors"
	"math"
)

// HTCKDomainSep is the domain separator used for the domain-separated hash
// function.
const HTCKDomainSep = "H_TCK"

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

// SignedReport contains a report and the corresponding signature. The client
// sends this to the server.
type SignedReport struct {
	Report *Report
	// This is a ed25519 signature in byte array form
	// The ed25519 package returns a byte array as the signature
	// here: https://golang.org/pkg/crypto/ed25519/#PrivateKey.Sign
	Sig []byte
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

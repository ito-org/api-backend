package tcn

import (
	"crypto/ed25519"
	"crypto/sha256"
	"errors"
	"fmt"
	"math"
)

// HTCKDomainSep is the domain separator used for the domain-separated hash
// function.
const HTCKDomainSep = "H_TCK"

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
	if _, err := nextHash.Write([]byte(HTCKDomainSep)); err != nil {
		fmt.Printf("Failed to write tck domain separator: %s\n", err.Error())
		return nil, err
	}
	if _, err := nextHash.Write(tck.RVK); err != nil {
		fmt.Printf("Failed to write rvk: %s\n", err.Error())
		return nil, err
	}
	if _, err := nextHash.Write(tck.TCKBytes[:]); err != nil {
		fmt.Printf("Failed to write tck bytes: %s\n", err.Error())
		return nil, err
	}

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

package main

import "crypto/ed25519"

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

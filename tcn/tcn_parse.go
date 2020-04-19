package tcn

import (
	"crypto/ed25519"
	"encoding/binary"
)

// GetSignedReport interprets data as a signed report and returns it as a
// parsed structure.
func GetSignedReport(data []byte) (*SignedReport, error) {
	tckBytes := [32]byte{}
	copy(tckBytes[:], data[32:64])

	memoDataLen := uint8(data[69])
	memo := &Memo{
		Type: data[68],
		Len:  memoDataLen,
		Data: data[70 : 70+memoDataLen],
	}

	report := &Report{
		RVK:      ed25519.PublicKey(data[:32]),
		TCKBytes: tckBytes,
		J1:       binary.LittleEndian.Uint16(data[64:66]),
		J2:       binary.LittleEndian.Uint16(data[66:68]),
		Memo:     memo,
	}

	sig := data[70+memoDataLen:]

	return &SignedReport{
		Report: report,
		Sig:    sig,
	}, nil
}

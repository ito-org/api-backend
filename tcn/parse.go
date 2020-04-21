package tcn

import (
	"crypto/ed25519"
	"encoding/binary"
	"errors"
)

// GetSignedReport interprets data as a signed report and returns it as a
// parsed structure.
func GetSignedReport(data []byte) (*SignedReport, error) {
	if len(data) < SignedReportMinLength+ed25519.SignatureSize {
		return nil, errors.New("Data too short to be a valid signed report")
	}

	report := GetReport(data)
	memoDataLen := uint8(data[69])
	sig := data[70+memoDataLen:]

	return &SignedReport{
		Report: report,
		Sig:    sig,
	}, nil
}

// GetReport inteprets data as a report and returns it as a parsed structure.
func GetReport(data []byte) *Report {
	_, report := getReport(data)
	return report
}

// getReport is the internal function for getting reports from byte arrays.
// It returns the report contained in the data field and also returns the
// length / end position of the array.
func getReport(data []byte) (endPos uint16, report *Report) {
	tckBytes := [32]byte{}
	copy(tckBytes[:], data[32:64])

	memoDataLen := uint8(data[69])

	memo := &Memo{
		Type: data[68],
		Len:  memoDataLen,
		Data: data[70 : 70+memoDataLen],
	}

	// TODO: do some array bounds checking

	report = &Report{
		RVK:      ed25519.PublicKey(data[:32]),
		TCKBytes: tckBytes,
		J1:       binary.LittleEndian.Uint16(data[64:66]),
		J2:       binary.LittleEndian.Uint16(data[66:68]),
		Memo:     memo,
	}

	endPos = 70 + uint16(memoDataLen)

	return endPos, report
}

// GetReports gets all reports contained in a byte array and returns them.
func GetReports(data []byte) []*Report {
	reports := []*Report{}
	var startPos uint16
	for true {
		endPos, report := getReport(data[startPos:])
		startPos += endPos
		reports = append(reports, report)
		if int(startPos) >= len(data) {
			// TODO: Fail if the startPos is greater than the length of data
			// because this can only happen if something was wrong with the
			// data (e.g. memo length field incorrect)
			break
		}
	}
	return reports
}

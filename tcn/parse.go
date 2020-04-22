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

	signedReport, _ := getSignedReport(data)
	return signedReport, nil
}

// getSignedReport returns the signed report contained in data and returns it
// in combination with its length (end position), which allows for parsing of
// multiple signed reports.
func getSignedReport(data []byte) (*SignedReport, uint16) {
	report, reportEndPos := getReport(data)
	endPos := reportEndPos + ed25519.SignatureSize
	sig := data[reportEndPos:endPos]
	return &SignedReport{
		Report: report,
		Sig:    sig,
	}, endPos
}

// GetSignedReports gets all signed reports contained in a byte array and
// returns them.
func GetSignedReports(data []byte) []*SignedReport {
	signedReports := []*SignedReport{}
	var startPos uint16
	for true {
		signedReport, endPos := getSignedReport(data[startPos:])
		startPos += endPos
		signedReports = append(signedReports, signedReport)
		if int(startPos) >= len(data) {
			// TODO: Fail if the startPos is greater than the length of data
			// because this can only happen if something was wrong with the
			// data (e.g. memo length field incorrect)
			break
		}
	}
	return signedReports
}

// GetReport inteprets data as a report and returns it as a parsed structure.
func GetReport(data []byte) *Report {
	report, _ := getReport(data)
	return report
}

// getReport is the internal function for getting reports from byte arrays.
// It returns the report contained in the data field and also returns the
// length / end position of the array.
func getReport(data []byte) (report *Report, endPos uint16) {
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

	return report, endPos
}

// GetReports gets all reports contained in a byte array and returns them.
func GetReports(data []byte) []*Report {
	reports := []*Report{}
	var startPos uint16
	for true {
		report, endPos := getReport(data[startPos:])
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

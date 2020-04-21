package tcn

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/binary"
	"errors"
)

const (
	// ITOMemoCode is the code that marks a report as an ito report in the
	// memo.
	ITOMemoCode = 0x2
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

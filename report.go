package main

// Report represents a report as described in the TCN protocol:
// https://github.com/TCNCoalition/TCN#reporting
type Report struct {
	RVK      string
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

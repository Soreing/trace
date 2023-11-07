package trace

import "encoding/hex"

// TraceInfo is a single data type containing trace id, parent id and span id.
type TraceInfo struct {
	tid [16]byte
	pid [8]byte
	sid [8]byte
}

// NewTraceInfo creates a TraceInfo object from trace id, parent id and span id.
func NewTraceInfo(
	tid [16]byte,
	pid [8]byte,
	sid [8]byte,
) *TraceInfo {
	inf := &TraceInfo{}
	copy(inf.tid[:], tid[:])
	copy(inf.pid[:], pid[:])
	copy(inf.sid[:], sid[:])
	return inf
}

// GetIds returns the trace id, parent id and span id as byte arrays.
func (inf *TraceInfo) GetIds() ([16]byte, [8]byte, [8]byte) {
	return inf.tid, inf.pid, inf.sid
}

// GetStringIds returns the trace id, parent id and span id as strings.
func (inf *TraceInfo) GetStringIds() (string, string, string) {
	tid := hex.EncodeToString(inf.tid[:])
	pid := hex.EncodeToString(inf.pid[:])
	sid := hex.EncodeToString(inf.sid[:])
	return tid, pid, sid
}

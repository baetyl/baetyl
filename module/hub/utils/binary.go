package utils

import (
	"encoding/binary"
)

// U64U64ToB converts two uint64 to bytes
func U64U64ToB(sid, ts uint64) []byte {
	r := make([]byte, 16)
	binary.BigEndian.PutUint64(r, sid)
	binary.BigEndian.PutUint64(r[8:], ts)
	return r
}

// U64U64 gets two uint64 from bytes
func U64U64(v []byte) (uint64, uint64) {
	return binary.BigEndian.Uint64(v), binary.BigEndian.Uint64(v[8:])
}

// U64ToB converts uint64 to bytes
func U64ToB(v uint64) []byte {
	r := make([]byte, 8)
	binary.BigEndian.PutUint64(r, v)
	return r
}

// U16 gets uint16 from bytes
func U16(v []byte) uint16 {
	return binary.BigEndian.Uint16(v)
}

// U64 gets uint64 from bytes
func U64(v []byte) uint64 {
	return binary.BigEndian.Uint64(v)
}

// PutU16 puts uint16 into bytes
func PutU16(dst []byte, v uint16) {
	binary.BigEndian.PutUint16(dst, v)
}

// PutU64 puts uint64 into bytes
func PutU64(dst []byte, v uint64) {
	binary.BigEndian.PutUint64(dst, v)
}

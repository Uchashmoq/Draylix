package drlx

import (
	"encoding/binary"
	"fmt"
)

func intToBytes(n int) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(n))
	return b
}
func int64ToBytes(n int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(n))
	return b
}

func bytesToInt(p []byte) int {
	u := binary.BigEndian.Uint32(p)
	return int(u)
}
func bytesToInt64(p []byte) int64 {
	u := binary.BigEndian.Uint64(p)
	return int64(u)
}

func BytesFormat(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float32(bytes)/1024.0)
	} else {
		return fmt.Sprintf("%.2f MB", float32(bytes)/1024.0/1024.0)
	}
}

package validation

import (
	"errors"
	"fmt"
	"hash/crc32"
	"time"
)

// ErrChecksumMismatch 批次校验和不匹配。
var ErrChecksumMismatch = errors.New("batch checksum mismatch")

// InWindow 判断 t 是否落在 [start,end] 内。
func InWindow(t, start, end time.Time) bool { return !t.Before(start) && !t.After(end) }

// Recent 判断 t 距 now 是否不超过 limit。
func Recent(t time.Time, limit time.Duration, now time.Time) bool { return now.Sub(t) <= limit }

// Window 是一个左闭右闭采集时间窗。
type Window struct {
	From time.Time
	To   time.Time
}

// NewWindow 构造时间窗。
func NewWindow(from, to time.Time) Window { return Window{From: from, To: to} }

// Contains 判断时间点是否在窗内。
func (w Window) Contains(t time.Time) bool { return InWindow(t, w.From, w.To) }

// Sign 对指纹串计算 CRC32 签名的十六进制表示。
func Sign(fingerprint string) string {
	return fmt.Sprintf("%08x", crc32.ChecksumIEEE([]byte(fingerprint)))
}

// VerifyChecksum 校验批次签名。
func VerifyChecksum(signature, fingerprint string) bool {
	return signature == Sign(fingerprint)
}

package storage

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"strings"
)

const (
	DefaultCopyBuffer = 256 * 1024
	sha256HexLen      = 64
)

// Integrity 分片级与聚合级完整性快照。
type Integrity struct {
	ChunkSize     int64    `json:"chunk_size"`
	ChunkCount    int      `json:"chunk_count"`
	ChunkSHA256   []string `json:"chunk_sha256"`
	MerkleSHA256  string   `json:"merkle_sha256"`
	ContentSHA256 string   `json:"content_sha256,omitempty"`
}

// NewSHA256 返回用于边写边算的哈希器。
func NewSHA256() hash.Hash {
	return sha256.New()
}

// HexSHA256 将摘要编码为小写 hex。
func HexSHA256(sum []byte) string {
	return hex.EncodeToString(sum)
}

// DecodeSHA256Hex 解析并校验 64 位 hex SHA-256。
func DecodeSHA256Hex(s string) ([]byte, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if len(s) != sha256HexLen {
		return nil, ErrChecksumMismatch
	}
	b, err := hex.DecodeString(s)
	if err != nil || len(b) != sha256.Size {
		return nil, ErrChecksumMismatch
	}
	return b, nil
}

// EqualSHA256Hex 常量表比较两个 hex 摘要。
func EqualSHA256Hex(a, b string) bool {
	aa, err1 := DecodeSHA256Hex(a)
	bb, err2 := DecodeSHA256Hex(b)
	if err1 != nil || err2 != nil {
		return false
	}
	if len(aa) != len(bb) {
		return false
	}
	var v byte
	for i := range aa {
		v |= aa[i] ^ bb[i]
	}
	return v == 0
}

// HashReader 边读边算 SHA-256，并强制精确读 size 字节。
func HashReader(r io.Reader, size int64, buf []byte) (hexDigest string, n int64, err error) {
	if size < 0 {
		return "", 0, ErrInvalidPart
	}
	if buf == nil {
		buf = make([]byte, DefaultCopyBuffer)
	}
	h := sha256.New()
	lr := io.LimitReader(r, size+1)
	written, err := io.CopyBuffer(h, lr, buf)
	if err != nil {
		return "", written, err
	}
	if written != size {
		return "", written, fmt.Errorf("%w: got %d want %d", ErrChecksumMismatch, written, size)
	}
	return HexSHA256(h.Sum(nil)), written, nil
}

// MerkleSHA256 按规范拼接 chunkSize、count 与各分片摘要后做 SHA-256（O(分片数)，不读正文）。
func MerkleSHA256(chunkSize int64, digests []string) (string, error) {
	// 1. 规范化各 leaf hex 为原始 32 字节
	leaves := make([][]byte, 0, len(digests))
	for _, d := range digests {
		b, err := DecodeSHA256Hex(d)
		if err != nil {
			return "", err
		}
		leaves = append(leaves, b)
	}
	// 2. canonical: magic | chunkSize(u64be) | count(u32be) | leaf0..leafN
	h := sha256.New()
	_, _ = h.Write([]byte("light-gozero/storage/merkle/v1\n"))
	var meta [12]byte
	binary.BigEndian.PutUint64(meta[0:8], uint64(chunkSize))
	binary.BigEndian.PutUint32(meta[8:12], uint32(len(leaves)))
	_, _ = h.Write(meta[:])
	for _, leaf := range leaves {
		_, _ = h.Write(leaf)
	}
	return HexSHA256(h.Sum(nil)), nil
}

// BuildIntegrity 由已校验分片摘要组装 Integrity。
func BuildIntegrity(chunkSize int64, digests []string, contentSHA256 string) (Integrity, error) {
	merkle, err := MerkleSHA256(chunkSize, digests)
	if err != nil {
		return Integrity{}, err
	}
	cp := append([]string(nil), digests...)
	return Integrity{
		ChunkSize:     chunkSize,
		ChunkCount:    len(cp),
		ChunkSHA256:   cp,
		MerkleSHA256:  merkle,
		ContentSHA256: contentSHA256,
	}, nil
}

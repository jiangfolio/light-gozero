package storage

import (
	"bytes"
	"strings"
	"testing"
)

func TestMerkleStableAndSensitive(t *testing.T) {
	d1 := strings.Repeat("a", 64)
	d2 := strings.Repeat("b", 64)
	// 非法 hex 长度
	if _, err := MerkleSHA256(8, []string{"abcd"}); err == nil {
		t.Fatal("want decode error")
	}
	m1, err := MerkleSHA256(8<<20, []string{d1, d2})
	if err != nil {
		t.Fatal(err)
	}
	m2, err := MerkleSHA256(8<<20, []string{d1, d2})
	if err != nil || m1 != m2 {
		t.Fatalf("unstable merkle")
	}
	m3, err := MerkleSHA256(8<<20, []string{d2, d1})
	if err != nil {
		t.Fatal(err)
	}
	if m1 == m3 {
		t.Fatal("order should change merkle")
	}
	m4, err := MerkleSHA256(4<<20, []string{d1, d2})
	if err != nil {
		t.Fatal(err)
	}
	if m1 == m4 {
		t.Fatal("chunkSize should change merkle")
	}
}

func TestHashReaderExact(t *testing.T) {
	payload := []byte("hello-storage")
	h := NewSHA256()
	_, _ = h.Write(payload)
	want := HexSHA256(h.Sum(nil))

	got, n, err := HashReader(bytes.NewReader(payload), int64(len(payload)), nil)
	if err != nil || n != int64(len(payload)) || got != want {
		t.Fatalf("got=%s n=%d err=%v want=%s", got, n, err, want)
	}
	if _, _, err := HashReader(bytes.NewReader(payload), int64(len(payload))-1, nil); err == nil {
		t.Fatal("want size mismatch when declaring smaller size with extra bytes available")
	}
}

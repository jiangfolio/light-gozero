package storage

import "testing"

func TestNormalizeKey(t *testing.T) {
	ok := []string{"a", "a/b", "tenant/obj/uuid-1", "A.B_c-1"}
	for _, k := range ok {
		got, err := NormalizeKey(k)
		if err != nil || got != k {
			t.Fatalf("NormalizeKey(%q)=%q,%v", k, got, err)
		}
	}
	bad := []string{
		"",
		"/abs",
		"..",
		"a/../b",
		"a//b",
		`a\b`,
		"C:/x",
		"a/./b",
		string([]byte{'a', 0x01, 'b'}),
	}
	for _, k := range bad {
		if _, err := NormalizeKey(k); err == nil {
			t.Fatalf("NormalizeKey(%q) want error", k)
		}
	}
	long := make([]byte, MaxKeyLen+1)
	for i := range long {
		long[i] = 'a'
	}
	if _, err := NormalizeKey(string(long)); err == nil {
		t.Fatal("expected too long key error")
	}
}

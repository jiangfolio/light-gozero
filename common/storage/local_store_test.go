package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"path/filepath"
	"strings"
	"testing"
)

func testLocalCfg(t *testing.T, access string) Config {
	t.Helper()
	root := t.TempDir()
	cfg := Config{
		Provider:            ProviderLocal,
		AccessMode:          access,
		ChunkSizeBytes:      64,
		MaxObjectBytes:      1024,
		UploadExpireSeconds: 3600,
		Local:               LocalConfig{Root: root},
	}
	if access == AccessPublic {
		cfg.PublicBaseURL = "https://api.example.com"
	}
	return cfg
}

func shaHex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func TestLocalUploadHappyPathAndPublicURL(t *testing.T) {
	ctx := context.Background()
	store, err := New(testLocalCfg(t, AccessPublic))
	if err != nil {
		t.Fatal(err)
	}
	key := "demo/hello.txt"
	p1 := bytes.Repeat([]byte("a"), 64)
	p2 := []byte("tail")
	id, err := store.Initiate(ctx, key, InitiateOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.PutPart(ctx, key, id, 1, bytes.NewReader(p1), int64(len(p1)), shaHex(p1)); err != nil {
		t.Fatal(err)
	}
	// 幂等重传
	if _, err := store.PutPart(ctx, key, id, 1, bytes.NewReader(p1), int64(len(p1)), shaHex(p1)); err != nil {
		t.Fatal(err)
	}
	if _, err := store.PutPart(ctx, key, id, 2, bytes.NewReader(p2), int64(len(p2)), shaHex(p2)); err != nil {
		t.Fatal(err)
	}
	integ, err := BuildIntegrity(64, []string{shaHex(p1), shaHex(p2)}, "")
	if err != nil {
		t.Fatal(err)
	}
	info, err := store.Complete(ctx, key, id, nil, integ.MerkleSHA256)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size != int64(len(p1)+len(p2)) {
		t.Fatalf("size=%d", info.Size)
	}
	rc, _, err := store.Open(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()
	all, _ := io.ReadAll(rc)
	if !bytes.Equal(all, append(p1, p2...)) {
		t.Fatal("content mismatch")
	}
	u, err := store.PublicURL(key)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(u, "https://api.example.com/api/v1/file/public?key=") {
		t.Fatalf("url=%s", u)
	}
	if strings.Contains(u, "X-Amz-") || strings.Contains(strings.ToLower(u), "expires=") {
		t.Fatalf("public url must not carry ttl/signature: %s", u)
	}
	// 物理路径应是哈希布局，而非原始 key
	ls := store.(*localStore)
	_, dataPath, _ := ls.objectPaths(key)
	if strings.Contains(dataPath, "demo") {
		t.Fatalf("path leaked key: %s", dataPath)
	}
	if filepath.Base(dataPath) != ObjectHash(key)+".bin" {
		t.Fatalf("unexpected object path %s", dataPath)
	}
}

func TestLocalChecksumAndConflict(t *testing.T) {
	ctx := context.Background()
	store, err := New(testLocalCfg(t, AccessPrivate))
	if err != nil {
		t.Fatal(err)
	}
	key := "k1"
	id, err := store.Initiate(ctx, key, InitiateOpts{})
	if err != nil {
		t.Fatal(err)
	}
	body := []byte("abcdefgh")
	if _, err := store.PutPart(ctx, key, id, 1, bytes.NewReader(body), int64(len(body)), shaHex([]byte("wrong!!!"))); !errors.Is(err, ErrChecksumMismatch) {
		t.Fatalf("want checksum mismatch, got %v", err)
	}
	if _, err := store.PutPart(ctx, key, id, 1, bytes.NewReader(body), int64(len(body)), shaHex(body)); err != nil {
		t.Fatal(err)
	}
	other := []byte("ABCDEFGH")
	if _, err := store.PutPart(ctx, key, id, 1, bytes.NewReader(other), int64(len(other)), shaHex(other)); !errors.Is(err, ErrPartExistsConflict) {
		t.Fatalf("want conflict, got %v", err)
	}
	if _, err := store.PresignGet(ctx, key, 0); !errors.Is(err, ErrNotSupported) {
		t.Fatalf("local private PresignGet: %v", err)
	}
	if _, err := store.PublicURL(key); !errors.Is(err, ErrNotSupported) {
		t.Fatalf("local private PublicURL: %v", err)
	}
}

func TestLocalRejectsPathTraversalKey(t *testing.T) {
	ctx := context.Background()
	store, err := New(testLocalCfg(t, AccessPrivate))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Initiate(ctx, "../etc/passwd", InitiateOpts{}); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("got %v", err)
	}
}

func TestLocalPartNumberAndSizeLimits(t *testing.T) {
	ctx := context.Background()
	store, err := New(testLocalCfg(t, AccessPrivate))
	if err != nil {
		t.Fatal(err)
	}
	key := "lim"
	id, _ := store.Initiate(ctx, key, InitiateOpts{})
	body := bytes.Repeat([]byte("x"), 65)
	if _, err := store.PutPart(ctx, key, id, 1, bytes.NewReader(body), 65, shaHex(body)); !errors.Is(err, ErrInvalidPart) {
		t.Fatalf("oversize part: %v", err)
	}
	if _, err := store.PutPart(ctx, key, id, 0, bytes.NewReader([]byte("a")), 1, shaHex([]byte("a"))); !errors.Is(err, ErrInvalidPart) {
		t.Fatalf("part 0: %v", err)
	}
}

func TestConfigRejectsLocalPublicWithoutBase(t *testing.T) {
	cfg := Config{
		Provider:   ProviderLocal,
		AccessMode: AccessPublic,
		Local:      LocalConfig{Root: t.TempDir()},
	}
	if _, err := New(cfg); err == nil {
		t.Fatal("want public base url error")
	}
}

func TestS3PublicURLShape(t *testing.T) {
	u, err := BuildS3PublicURL("https://cdn.example.com/bucket", "a/b")
	if err != nil {
		t.Fatal(err)
	}
	if u != "https://cdn.example.com/bucket/a/b" {
		t.Fatalf("got %s", u)
	}
}

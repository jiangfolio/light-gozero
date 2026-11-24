package storage

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestS3AccessModeAndDirectUpload(t *testing.T) {
	priv := Config{
		Provider:       ProviderS3,
		AccessMode:     AccessPrivate,
		ChunkSizeBytes: MinS3ChunkSizeBytes,
		MaxObjectBytes: 32 << 20,
		S3: S3Config{
			Bucket:    "light-gozero",
			AccessKey: "ak",
			SecretKey: "sk",
			Region:    "us-east-1",
		},
	}
	store, err := New(priv)
	if err != nil {
		t.Fatal(err)
	}
	if !store.Capabilities().DirectUpload || store.Capabilities().ServerPutPart {
		t.Fatalf("caps=%+v", store.Capabilities())
	}
	if _, err := store.PutPart(context.Background(), "a/b", "upload", 1, strings.NewReader("x"), 1, strings.Repeat("a", 64)); !errors.Is(err, ErrDirectUploadOnly) {
		t.Fatalf("PutPart: %v", err)
	}
	if _, err := store.PublicURL("a/b"); !errors.Is(err, ErrNotSupported) {
		t.Fatalf("private PublicURL: %v", err)
	}

	pub := priv
	pub.AccessMode = AccessPublic
	pub.PublicBaseURL = "https://cdn.example.com/light"
	store2, err := New(pub)
	if err != nil {
		t.Fatal(err)
	}
	u, err := store2.PublicURL("a/b")
	if err != nil {
		t.Fatal(err)
	}
	if u != "https://cdn.example.com/light/a/b" {
		t.Fatalf("url=%s", u)
	}
	if _, err := store2.PresignGet(context.Background(), "a/b", 0); !errors.Is(err, ErrNotSupported) {
		t.Fatalf("public PresignGet: %v", err)
	}
}

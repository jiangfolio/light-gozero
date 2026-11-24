package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	uploadStatusUploading  = "uploading"
	uploadStatusCompleting = "completing"
	uploadStatusCompleted  = "completed"
	uploadStatusAborted    = "aborted"
)

type localStore struct {
	cfg        Config
	root       string
	chunkSize  int64
	maxObject  int64
	maxParts   int
	expire     time.Duration
	fullHash   bool
	mu         sync.Mutex
	uploadMu   map[string]*sync.Mutex
}

type uploadMeta struct {
	UploadID   string                `json:"upload_id"`
	Key        string                `json:"key"`
	Status     string                `json:"status"`
	ChunkSize  int64                 `json:"chunk_size"`
	CreatedAt  time.Time             `json:"created_at"`
	ExpireAt   time.Time             `json:"expire_at"`
	Parts      map[string]PartResult `json:"parts"` // partNumber string -> result
	ContentType string               `json:"content_type,omitempty"`
}

type objectSidecar struct {
	Key       string    `json:"key"`
	Size      int64     `json:"size"`
	Integrity Integrity `json:"integrity"`
}

func newLocalStore(cfg Config) (Store, error) {
	// 1. 解析 Root 为绝对路径并拒绝危险宽路径
	root, err := filepath.Abs(strings.TrimSpace(cfg.Local.Root))
	if err != nil {
		return nil, fmt.Errorf("%w: resolve root: %v", ErrInvalidConfig, err)
	}
	if err := assertSafeRoot(root); err != nil {
		return nil, err
	}
	// 2. 创建 Root 并收紧权限
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, fmt.Errorf("%w: mkdir root: %v", ErrInvalidConfig, err)
	}
	if err := os.Chmod(root, 0o700); err != nil {
		return nil, fmt.Errorf("%w: chmod root: %v", ErrInvalidConfig, err)
	}
	if err := ensureUnderRoot(root, root); err != nil {
		return nil, err
	}
	return &localStore{
		cfg:       cfg,
		root:      root,
		chunkSize: cfg.ChunkSizeBytes,
		maxObject: cfg.MaxObjectBytes,
		maxParts:  cfg.MaxParts(),
		expire:    time.Duration(cfg.UploadExpireSeconds) * time.Second,
		fullHash:  cfg.Integrity.FullContentHash,
		uploadMu:  make(map[string]*sync.Mutex),
	}, nil
}

func assertSafeRoot(root string) error {
	clean := filepath.Clean(root)
	if clean == "/" || clean == "." || clean == "" {
		return fmt.Errorf("%w: dangerous Local.Root", ErrInvalidConfig)
	}
	// 拒绝过短路径（如 /tmp 仍可用，但拒绝单段系统根下直接用家目录名过于宽的由运维约束）
	return nil
}

func (s *localStore) Provider() string   { return ProviderLocal }
func (s *localStore) AccessMode() string { return s.cfg.AccessMode }
func (s *localStore) ChunkSize() int64   { return s.chunkSize }
func (s *localStore) MaxObjectBytes() int64 { return s.maxObject }
func (s *localStore) MaxParts() int      { return s.maxParts }

func (s *localStore) Capabilities() Caps {
	return Caps{
		DirectUpload:  false,
		ServerPutPart: true,
		PresignGet:    false,
		PublicURL:     s.cfg.AccessMode == AccessPublic,
	}
}

func (s *localStore) lockUpload(id string) func() {
	s.mu.Lock()
	m, ok := s.uploadMu[id]
	if !ok {
		m = &sync.Mutex{}
		s.uploadMu[id] = m
	}
	s.mu.Unlock()
	m.Lock()
	return m.Unlock
}

func (s *localStore) uploadsDir() string { return filepath.Join(s.root, "uploads") }
func (s *localStore) objectsDir() string { return filepath.Join(s.root, "objects") }

func (s *localStore) uploadDir(id string) string {
	return filepath.Join(s.uploadsDir(), id)
}

func (s *localStore) objectPaths(key string) (dir, data, side string) {
	h := ObjectHash(key)
	dir = filepath.Join(s.objectsDir(), h[:2])
	data = filepath.Join(dir, h+".bin")
	side = filepath.Join(dir, h+".integrity.json")
	return dir, data, side
}

// Initiate 创建分片上传会话。
func (s *localStore) Initiate(ctx context.Context, key string, opts InitiateOpts) (string, error) {
	// 1. 规范化 key 与空间预检
	key, err := NormalizeKey(key)
	if err != nil {
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	maxSize := opts.MaxSize
	if maxSize <= 0 {
		maxSize = s.maxObject
	}
	if maxSize > s.maxObject {
		return "", ErrTooLarge
	}
	if err := s.ensureDiskSpace(maxSize); err != nil {
		return "", err
	}
	// 2. 生成高熵 uploadID 并写 meta
	id, err := newUploadID()
	if err != nil {
		return "", err
	}
	now := time.Now().UTC()
	meta := uploadMeta{
		UploadID:    id,
		Key:         key,
		Status:      uploadStatusUploading,
		ChunkSize:   s.chunkSize,
		CreatedAt:   now,
		ExpireAt:    now.Add(s.expire),
		Parts:       map[string]PartResult{},
		ContentType: opts.ContentType,
	}
	dir := s.uploadDir(id)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	if err := s.writeMeta(dir, &meta); err != nil {
		_ = os.RemoveAll(dir)
		return "", err
	}
	return id, nil
}

func (s *localStore) PreparePart(ctx context.Context, key, uploadID string, partNumber int, opts PartOpts) (PartPrep, error) {
	// local 由调用方直接 PutPart，无需预签名
	if _, _, err := s.loadUpload(key, uploadID); err != nil {
		return PartPrep{}, err
	}
	if err := ValidatePartNumber(partNumber, s.maxParts); err != nil {
		return PartPrep{}, err
	}
	return PartPrep{}, nil
}

// PutPart 流式写入分片并边写边校验 SHA-256。
func (s *localStore) PutPart(ctx context.Context, key, uploadID string, partNumber int, r io.Reader, size int64, chunkSHA256 string) (PartResult, error) {
	key, err := NormalizeKey(key)
	if err != nil {
		return PartResult{}, err
	}
	if err := ValidatePartNumber(partNumber, s.maxParts); err != nil {
		return PartResult{}, err
	}
	if err := ValidatePartSize(size, s.chunkSize, s.maxObject); err != nil {
		return PartResult{}, err
	}
	if _, err := DecodeSHA256Hex(chunkSHA256); err != nil {
		return PartResult{}, ErrChecksumMismatch
	}

	unlock := s.lockUpload(uploadID)
	defer unlock()

	dir, meta, err := s.loadUpload(key, uploadID)
	if err != nil {
		return PartResult{}, err
	}
	if meta.Status != uploadStatusUploading {
		return PartResult{}, ErrUploadState
	}
	if time.Now().UTC().After(meta.ExpireAt) {
		return PartResult{}, ErrUploadExpired
	}

	// 1. 幂等：同分片同哈希同大小直接成功
	if old, ok := meta.Parts[strconv.Itoa(partNumber)]; ok {
		if old.Size == size && EqualSHA256Hex(old.SHA256, chunkSHA256) {
			return old, nil
		}
		return PartResult{}, ErrPartExistsConflict
	}

	if err := s.ensureDiskSpace(size); err != nil {
		return PartResult{}, err
	}

	// 2. 原子写 tmp → 边写边哈希 → rename
	partName := fmt.Sprintf("part_%d.bin", partNumber)
	finalPath := filepath.Join(dir, partName)
	tmpPath := finalPath + ".tmp"
	_ = os.Remove(tmpPath)

	f, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o600)
	if err != nil {
		return PartResult{}, err
	}
	h := NewSHA256()
	buf := make([]byte, DefaultCopyBuffer)
	lr := io.LimitReader(r, size+1)
	mw := io.MultiWriter(f, h)

	written, copyErr := copyWithContext(ctx, mw, lr, buf)
	closeErr := f.Close()
	if copyErr != nil || closeErr != nil || written != size {
		_ = os.Remove(tmpPath)
		if copyErr != nil {
			return PartResult{}, copyErr
		}
		if closeErr != nil {
			return PartResult{}, closeErr
		}
		return PartResult{}, fmt.Errorf("%w: size mismatch", ErrChecksumMismatch)
	}
	got := HexSHA256(h.Sum(nil))
	if !EqualSHA256Hex(got, chunkSHA256) {
		_ = os.Remove(tmpPath)
		return PartResult{}, ErrChecksumMismatch
	}
	if err := os.Rename(tmpPath, finalPath); err != nil {
		_ = os.Remove(tmpPath)
		return PartResult{}, err
	}
	_ = os.Chmod(finalPath, 0o600)

	res := PartResult{PartNumber: partNumber, Size: size, SHA256: got}
	meta.Parts[strconv.Itoa(partNumber)] = res
	if err := s.writeMeta(dir, meta); err != nil {
		return PartResult{}, err
	}
	return res, nil
}

func (s *localStore) ListParts(ctx context.Context, key, uploadID string) ([]PartResult, error) {
	_, meta, err := s.loadUpload(key, uploadID)
	if err != nil {
		return nil, err
	}
	out := make([]PartResult, 0, len(meta.Parts))
	for _, p := range meta.Parts {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].PartNumber < out[j].PartNumber })
	return out, nil
}

// Complete 校验分片齐全、聚合 Merkle、流式合并为最终对象。
func (s *localStore) Complete(ctx context.Context, key, uploadID string, parts []PartResult, expectMerkle string) (ObjectInfo, error) {
	key, err := NormalizeKey(key)
	if err != nil {
		return ObjectInfo{}, err
	}
	unlock := s.lockUpload(uploadID)
	defer unlock()

	dir, meta, err := s.loadUpload(key, uploadID)
	if err != nil {
		return ObjectInfo{}, err
	}
	if meta.Status != uploadStatusUploading && meta.Status != uploadStatusCompleting {
		return ObjectInfo{}, ErrUploadState
	}
	if time.Now().UTC().After(meta.ExpireAt) {
		return ObjectInfo{}, ErrUploadExpired
	}

	// 1. 以服务端已存分片为准组装有序摘要（忽略客户端伪造哈希）
	if len(meta.Parts) == 0 {
		return ObjectInfo{}, ErrInvalidPart
	}
	ordered := make([]PartResult, 0, len(meta.Parts))
	for _, p := range meta.Parts {
		ordered = append(ordered, p)
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].PartNumber < ordered[j].PartNumber })
	for i, p := range ordered {
		wantNo := i + 1
		if p.PartNumber != wantNo {
			return ObjectInfo{}, fmt.Errorf("%w: missing part %d", ErrInvalidPart, wantNo)
		}
		if i < len(ordered)-1 && p.Size != s.chunkSize {
			return ObjectInfo{}, fmt.Errorf("%w: non-final part size", ErrInvalidPart)
		}
	}
	var total int64
	digests := make([]string, 0, len(ordered))
	for _, p := range ordered {
		total += p.Size
		digests = append(digests, p.SHA256)
	}
	if total > s.maxObject {
		return ObjectInfo{}, ErrTooLarge
	}
	integ, err := BuildIntegrity(s.chunkSize, digests, "")
	if err != nil {
		return ObjectInfo{}, err
	}
	if expectMerkle != "" && !EqualSHA256Hex(expectMerkle, integ.MerkleSHA256) {
		return ObjectInfo{}, ErrMerkleMismatch
	}

	// 2. 进入 completing，合并到对象 tmp
	meta.Status = uploadStatusCompleting
	if err := s.writeMeta(dir, meta); err != nil {
		return ObjectInfo{}, err
	}
	objDir, dataPath, sidePath := s.objectPaths(key)
	if err := os.MkdirAll(objDir, 0o700); err != nil {
		return ObjectInfo{}, err
	}
	if err := ensureUnderRoot(s.root, objDir); err != nil {
		return ObjectInfo{}, err
	}
	tmpData := dataPath + ".tmp"
	tmpSide := sidePath + ".tmp"
	_ = os.Remove(tmpData)
	_ = os.Remove(tmpSide)

	out, err := os.OpenFile(tmpData, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o600)
	if err != nil {
		return ObjectInfo{}, err
	}
	var contentH = NewSHA256()
	var writer io.Writer = out
	if s.fullHash {
		writer = io.MultiWriter(out, contentH)
	}
	buf := make([]byte, DefaultCopyBuffer)
	for _, p := range ordered {
		if err := ctx.Err(); err != nil {
			_ = out.Close()
			_ = os.Remove(tmpData)
			return ObjectInfo{}, err
		}
		partPath := filepath.Join(dir, fmt.Sprintf("part_%d.bin", p.PartNumber))
		if err := assertRegularFileUnderRoot(s.root, partPath); err != nil {
			_ = out.Close()
			_ = os.Remove(tmpData)
			return ObjectInfo{}, err
		}
		in, err := os.Open(partPath)
		if err != nil {
			_ = out.Close()
			_ = os.Remove(tmpData)
			return ObjectInfo{}, err
		}
		_, copyErr := io.CopyBuffer(writer, io.LimitReader(in, p.Size), buf)
		_ = in.Close()
		if copyErr != nil {
			_ = out.Close()
			_ = os.Remove(tmpData)
			return ObjectInfo{}, copyErr
		}
	}
	if err := out.Sync(); err != nil {
		_ = out.Close()
		_ = os.Remove(tmpData)
		return ObjectInfo{}, err
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(tmpData)
		return ObjectInfo{}, err
	}
	if s.fullHash {
		integ.ContentSHA256 = HexSHA256(contentH.Sum(nil))
	}
	side := objectSidecar{Key: key, Size: total, Integrity: integ}
	sideBytes, err := json.MarshalIndent(side, "", "  ")
	if err != nil {
		_ = os.Remove(tmpData)
		return ObjectInfo{}, err
	}
	if err := os.WriteFile(tmpSide, sideBytes, 0o600); err != nil {
		_ = os.Remove(tmpData)
		return ObjectInfo{}, err
	}

	// 3. rename 原子可见后清理上传目录
	if err := os.Rename(tmpData, dataPath); err != nil {
		_ = os.Remove(tmpData)
		_ = os.Remove(tmpSide)
		return ObjectInfo{}, err
	}
	if err := os.Rename(tmpSide, sidePath); err != nil {
		_ = os.Remove(tmpSide)
		return ObjectInfo{}, err
	}
	_ = os.Chmod(dataPath, 0o600)
	_ = os.Chmod(sidePath, 0o600)
	meta.Status = uploadStatusCompleted
	_ = s.writeMeta(dir, meta)
	_ = os.RemoveAll(dir)

	return ObjectInfo{Key: key, Size: total, Integrity: integ}, nil
}

func (s *localStore) Abort(ctx context.Context, key, uploadID string) error {
	key, err := NormalizeKey(key)
	if err != nil {
		return err
	}
	unlock := s.lockUpload(uploadID)
	defer unlock()
	dir := s.uploadDir(uploadID)
	meta, err := s.readMeta(dir)
	if err != nil {
		if errors.Is(err, ErrUploadNotFound) {
			return nil
		}
		return err
	}
	if meta.Key != key {
		return ErrUploadNotFound
	}
	_ = os.RemoveAll(dir)
	return nil
}

func (s *localStore) Head(ctx context.Context, key string) (ObjectInfo, error) {
	key, err := NormalizeKey(key)
	if err != nil {
		return ObjectInfo{}, err
	}
	_, dataPath, sidePath := s.objectPaths(key)
	if err := assertRegularFileUnderRoot(s.root, dataPath); err != nil {
		return ObjectInfo{}, ErrObjectNotFound
	}
	st, err := os.Lstat(dataPath)
	if err != nil || !st.Mode().IsRegular() {
		return ObjectInfo{}, ErrObjectNotFound
	}
	side, err := readSidecar(sidePath)
	if err != nil {
		return ObjectInfo{}, err
	}
	return ObjectInfo{Key: key, Size: st.Size(), Integrity: side.Integrity}, nil
}

func (s *localStore) Open(ctx context.Context, key string) (io.ReadCloser, ObjectInfo, error) {
	info, err := s.Head(ctx, key)
	if err != nil {
		return nil, ObjectInfo{}, err
	}
	_, dataPath, _ := s.objectPaths(key)
	if err := assertRegularFileUnderRoot(s.root, dataPath); err != nil {
		return nil, ObjectInfo{}, err
	}
	f, err := os.Open(dataPath)
	if err != nil {
		return nil, ObjectInfo{}, ErrObjectNotFound
	}
	return f, info, nil
}

func (s *localStore) PresignGet(ctx context.Context, key string, expire time.Duration) (string, error) {
	return "", ErrNotSupported
}

func (s *localStore) PublicURL(key string) (string, error) {
	if s.cfg.AccessMode != AccessPublic {
		return "", ErrNotSupported
	}
	return BuildLocalPublicURL(s.cfg.PublicBaseURL, key)
}

func (s *localStore) Delete(ctx context.Context, key string) error {
	key, err := NormalizeKey(key)
	if err != nil {
		return err
	}
	_, dataPath, sidePath := s.objectPaths(key)
	_ = os.Remove(dataPath)
	_ = os.Remove(sidePath)
	return nil
}

func (s *localStore) GCExpired(ctx context.Context, now time.Time) (int, error) {
	uploads := s.uploadsDir()
	entries, err := os.ReadDir(uploads)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	removed := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if err := ctx.Err(); err != nil {
			return removed, err
		}
		dir := filepath.Join(uploads, e.Name())
		meta, err := s.readMeta(dir)
		if err != nil {
			continue
		}
		if now.UTC().After(meta.ExpireAt) && meta.Status != uploadStatusCompleted {
			if err := os.RemoveAll(dir); err == nil {
				removed++
			}
		}
	}
	return removed, nil
}

func (s *localStore) loadUpload(key, uploadID string) (string, *uploadMeta, error) {
	key, err := NormalizeKey(key)
	if err != nil {
		return "", nil, err
	}
	if err := validateUploadID(uploadID); err != nil {
		return "", nil, err
	}
	dir := s.uploadDir(uploadID)
	meta, err := s.readMeta(dir)
	if err != nil {
		return "", nil, err
	}
	if meta.Key != key {
		return "", nil, ErrUploadNotFound
	}
	return dir, meta, nil
}

func (s *localStore) readMeta(dir string) (*uploadMeta, error) {
	path := filepath.Join(dir, "meta.json")
	if err := assertRegularFileUnderRoot(s.root, path); err != nil {
		return nil, ErrUploadNotFound
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrUploadNotFound
		}
		return nil, err
	}
	var meta uploadMeta
	if err := json.Unmarshal(b, &meta); err != nil {
		return nil, err
	}
	if meta.Parts == nil {
		meta.Parts = map[string]PartResult{}
	}
	return &meta, nil
}

func (s *localStore) writeMeta(dir string, meta *uploadMeta) error {
	path := filepath.Join(dir, "meta.json")
	tmp := path + ".tmp"
	b, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func (s *localStore) ensureDiskSpace(need int64) error {
	var st syscall.Statfs_t
	if err := syscall.Statfs(s.root, &st); err != nil {
		return nil // 无法探测时不阻断
	}
	free := int64(st.Bavail) * int64(st.Bsize)
	// 预留一倍 need 作合并峰值粗估下限
	if free < need {
		return ErrInsufficientSpace
	}
	return nil
}

func readSidecar(path string) (objectSidecar, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return objectSidecar{}, ErrObjectNotFound
	}
	var side objectSidecar
	if err := json.Unmarshal(b, &side); err != nil {
		return objectSidecar{}, err
	}
	return side, nil
}

func newUploadID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

func validateUploadID(id string) error {
	id = strings.TrimSpace(id)
	if len(id) != 32 {
		return ErrUploadNotFound
	}
	if _, err := hex.DecodeString(id); err != nil {
		return ErrUploadNotFound
	}
	return nil
}

func ensureUnderRoot(root, path string) error {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	// 尽量解析 symlink 后再比前缀
	if evalRoot, err := filepath.EvalSymlinks(absRoot); err == nil {
		absRoot = evalRoot
	}
	if evalPath, err := filepath.EvalSymlinks(absPath); err == nil {
		absPath = evalPath
	}
	rel, err := filepath.Rel(absRoot, absPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return ErrInvalidKey
	}
	return nil
}

func assertRegularFileUnderRoot(root, path string) error {
	if err := ensureUnderRoot(root, filepath.Dir(path)); err != nil {
		return err
	}
	st, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if st.Mode()&os.ModeSymlink != 0 {
		return ErrInvalidKey
	}
	if !st.Mode().IsRegular() && !st.IsDir() {
		return ErrInvalidKey
	}
	return nil
}

func copyWithContext(ctx context.Context, dst io.Writer, src io.Reader, buf []byte) (int64, error) {
	if buf == nil {
		buf = make([]byte, DefaultCopyBuffer)
	}
	var written int64
	for {
		if err := ctx.Err(); err != nil {
			return written, err
		}
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				return written, ew
			}
			if nr != nw {
				return written, io.ErrShortWrite
			}
		}
		if er != nil {
			if er == io.EOF {
				return written, nil
			}
			return written, er
		}
	}
}

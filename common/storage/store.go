package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"
)

const (
	ProviderLocal = "local"
	ProviderS3    = "s3"

	AccessPrivate = "private"
	AccessPublic  = "public"

	DefaultChunkSizeBytes     = 8 << 20 // 8MiB
	DefaultMaxObjectBytes     = 32 << 30 // 32GiB
	DefaultUploadExpireSec    = 86400
	DefaultPresignExpireSec   = 900
	DefaultMaxPartsCap        = 10000
	MinS3ChunkSizeBytes       = 5 << 20 // S3 非最后分片下限
	LocalPublicAPIPathPrefix  = "/api/v1/file/public"
)

// Caps 描述驱动能力，供上层按能力组响应字段。
type Caps struct {
	DirectUpload  bool
	ServerPutPart bool
	PresignGet    bool
	PublicURL     bool
}

// InitiateOpts 创建上传会话选项。
type InitiateOpts struct {
	ContentType string
	MaxSize     int64 // 0 表示用配置 MaxObjectBytes
}

// PartOpts 预签名/准备分片选项。
type PartOpts struct {
	ChunkSHA256 string
	Size        int64
}

// PartPrep 分片准备结果；local 可为空，s3 含预签名。
type PartPrep struct {
	UploadURL string
	Headers   map[string]string
	ExpiresAt time.Time
}

// PartResult 已确认分片。
type PartResult struct {
	PartNumber  int
	Size        int64
	SHA256      string
	ETag        string
}

// ObjectInfo 已完成对象元信息。
type ObjectInfo struct {
	Key       string
	Size      int64
	Integrity Integrity
}

// Store 统一对象存储抽象（local 分片 / s3 直传语义对齐）。
type Store interface {
	Provider() string
	AccessMode() string
	Capabilities() Caps
	ChunkSize() int64
	MaxObjectBytes() int64
	MaxParts() int

	Initiate(ctx context.Context, key string, opts InitiateOpts) (uploadID string, err error)
	PreparePart(ctx context.Context, key, uploadID string, partNumber int, opts PartOpts) (PartPrep, error)
	PutPart(ctx context.Context, key, uploadID string, partNumber int, r io.Reader, size int64, chunkSHA256 string) (PartResult, error)
	ListParts(ctx context.Context, key, uploadID string) ([]PartResult, error)
	Complete(ctx context.Context, key, uploadID string, parts []PartResult, expectMerkle string) (ObjectInfo, error)
	Abort(ctx context.Context, key, uploadID string) error
	Head(ctx context.Context, key string) (ObjectInfo, error)
	Open(ctx context.Context, key string) (io.ReadCloser, ObjectInfo, error)
	PresignGet(ctx context.Context, key string, expire time.Duration) (string, error)
	PublicURL(key string) (string, error)
	Delete(ctx context.Context, key string) error
	GCExpired(ctx context.Context, now time.Time) (removed int, err error)
}

// Config yaml 可映射的存储配置。
type Config struct {
	Provider              string
	AccessMode            string `json:",default=private"`
	PublicBaseURL         string `json:",optional"`
	AllowInsecurePublicURL bool  `json:",optional"`
	ChunkSizeBytes        int64  `json:",default=8388608"`
	MaxObjectBytes        int64  `json:",default=34359738368"`
	UploadExpireSeconds   int64  `json:",default=86400"`
	Integrity             struct {
		FullContentHash bool `json:",optional"`
	} `json:",optional"`
	Local LocalConfig `json:",optional"`
	S3    S3Config    `json:",optional"`
}

// LocalConfig 本地盘配置。
type LocalConfig struct {
	Root string
}

// S3Config S3 兼容配置。
type S3Config struct {
	Endpoint             string `json:",optional"`
	Region               string `json:",default=us-east-1"`
	Bucket               string
	AccessKey            string `json:",env=SYSTEM_STORAGE_S3_ACCESS_KEY"`
	SecretKey            string `json:",env=SYSTEM_STORAGE_S3_SECRET_KEY"`
	ForcePathStyle       bool   `json:",default=true"`
	PresignExpireSeconds int64  `json:",default=900"`
}

// Validate 校验配置并填充默认值副作用到副本逻辑（调用方可先 Normalize）。
func (c *Config) NormalizeAndValidate() error {
	// 1. Provider / AccessMode 默认与枚举
	c.Provider = strings.ToLower(strings.TrimSpace(c.Provider))
	if c.Provider == "" {
		c.Provider = ProviderLocal
	}
	if c.Provider != ProviderLocal && c.Provider != ProviderS3 {
		return fmt.Errorf("%w: provider %q", ErrInvalidConfig, c.Provider)
	}
	c.AccessMode = strings.ToLower(strings.TrimSpace(c.AccessMode))
	if c.AccessMode == "" {
		c.AccessMode = AccessPrivate
	}
	if c.AccessMode != AccessPrivate && c.AccessMode != AccessPublic {
		return fmt.Errorf("%w: access_mode %q", ErrInvalidConfig, c.AccessMode)
	}

	// 2. 分片与对象硬顶
	if c.ChunkSizeBytes <= 0 {
		c.ChunkSizeBytes = DefaultChunkSizeBytes
	}
	if c.MaxObjectBytes <= 0 {
		c.MaxObjectBytes = DefaultMaxObjectBytes
	}
	if c.ChunkSizeBytes > c.MaxObjectBytes {
		return fmt.Errorf("%w: chunk size > max object", ErrInvalidConfig)
	}
	if c.Provider == ProviderS3 && c.ChunkSizeBytes < MinS3ChunkSizeBytes {
		return fmt.Errorf("%w: s3 chunk size must be >= 5MiB", ErrInvalidConfig)
	}
	if c.UploadExpireSeconds <= 0 {
		c.UploadExpireSeconds = DefaultUploadExpireSec
	}

	// 3. 公开读要求 PublicBaseURL
	if c.AccessMode == AccessPublic {
		base := strings.TrimSpace(c.PublicBaseURL)
		if base == "" {
			return fmt.Errorf("%w: PublicBaseURL required when AccessMode=public", ErrInvalidConfig)
		}
		u, err := url.Parse(base)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("%w: invalid PublicBaseURL", ErrInvalidConfig)
		}
		if u.Scheme != "https" && !c.AllowInsecurePublicURL {
			return fmt.Errorf("%w: PublicBaseURL must be https", ErrInvalidConfig)
		}
		c.PublicBaseURL = strings.TrimRight(base, "/")
	}

	// 4. Provider 专属
	switch c.Provider {
	case ProviderLocal:
		if strings.TrimSpace(c.Local.Root) == "" {
			return fmt.Errorf("%w: Local.Root required", ErrInvalidConfig)
		}
	case ProviderS3:
		if strings.TrimSpace(c.S3.Bucket) == "" {
			return fmt.Errorf("%w: S3.Bucket required", ErrInvalidConfig)
		}
		if strings.TrimSpace(c.S3.AccessKey) == "" || strings.TrimSpace(c.S3.SecretKey) == "" {
			return fmt.Errorf("%w: S3 credentials required", ErrInvalidConfig)
		}
		if c.S3.Region == "" {
			c.S3.Region = "us-east-1"
		}
		if c.S3.PresignExpireSeconds <= 0 {
			c.S3.PresignExpireSeconds = DefaultPresignExpireSec
		}
	}
	return nil
}

// MaxParts 由对象硬顶与分片大小推导，并封顶。
func (c Config) MaxParts() int {
	if c.ChunkSizeBytes <= 0 {
		return 0
	}
	n := int((c.MaxObjectBytes + c.ChunkSizeBytes - 1) / c.ChunkSizeBytes)
	if n > DefaultMaxPartsCap {
		return DefaultMaxPartsCap
	}
	if n < 1 {
		return 1
	}
	return n
}

// BuildLocalPublicURL 拼稳定公开 API 直链（无 TTL/签名参数）。
func BuildLocalPublicURL(publicBaseURL, key string) (string, error) {
	key, err := NormalizeKey(key)
	if err != nil {
		return "", err
	}
	base := strings.TrimRight(strings.TrimSpace(publicBaseURL), "/")
	if base == "" {
		return "", ErrNotSupported
	}
	return base + LocalPublicAPIPathPrefix + "?key=" + url.QueryEscape(key), nil
}

// BuildS3PublicURL 拼 CDN/桶公网无签名 URL。
func BuildS3PublicURL(publicBaseURL, key string) (string, error) {
	key, err := NormalizeKey(key)
	if err != nil {
		return "", err
	}
	base := strings.TrimRight(strings.TrimSpace(publicBaseURL), "/")
	if base == "" {
		return "", ErrNotSupported
	}
	return base + "/" + pathEscapeKey(key), nil
}

func pathEscapeKey(key string) string {
	parts := strings.Split(key, "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	return strings.Join(parts, "/")
}

// New 按配置构造 Store。
func New(cfg Config) (Store, error) {
	if err := cfg.NormalizeAndValidate(); err != nil {
		return nil, err
	}
	switch cfg.Provider {
	case ProviderLocal:
		return newLocalStore(cfg)
	case ProviderS3:
		return newS3Store(cfg)
	default:
		return nil, fmt.Errorf("%w: provider %q", ErrInvalidConfig, cfg.Provider)
	}
}

// ValidatePartNumber 校验分片号范围。
func ValidatePartNumber(partNumber, maxParts int) error {
	if partNumber < 1 || partNumber > maxParts {
		return ErrInvalidPart
	}
	return nil
}

// ValidatePartSize 校验非最后/最后分片大小约定；isLast 由调用方在 Complete 时判定，Put 时仅校验 1..chunkSize。
func ValidatePartSize(size, chunkSize, maxObject int64) error {
	if size <= 0 || size > chunkSize {
		return ErrInvalidPart
	}
	if size > maxObject {
		return ErrTooLarge
	}
	return nil
}

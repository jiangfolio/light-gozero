package storage

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type s3Store struct {
	cfg       Config
	client    *s3.Client
	presign   *s3.PresignClient
	bucket    string
	chunkSize int64
	maxObject int64
	maxParts  int
	presignTTL time.Duration
}

func newS3Store(cfg Config) (Store, error) {
	// 1. 组装静态凭证与可选自定义 Endpoint（MinIO 等）
	creds := credentials.NewStaticCredentialsProvider(cfg.S3.AccessKey, cfg.S3.SecretKey, "")
	endpoint := strings.TrimSpace(cfg.S3.Endpoint)
	client := s3.New(s3.Options{
		Region:       cfg.S3.Region,
		Credentials:  creds,
		BaseEndpoint: awsStringOrNil(endpoint),
		UsePathStyle: cfg.S3.ForcePathStyle,
		HTTPClient:   http.DefaultClient,
	})
	return &s3Store{
		cfg:        cfg,
		client:     client,
		presign:    s3.NewPresignClient(client),
		bucket:     cfg.S3.Bucket,
		chunkSize:  cfg.ChunkSizeBytes,
		maxObject:  cfg.MaxObjectBytes,
		maxParts:   cfg.MaxParts(),
		presignTTL: time.Duration(cfg.S3.PresignExpireSeconds) * time.Second,
	}, nil
}

func awsStringOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return aws.String(s)
}

func (s *s3Store) Provider() string      { return ProviderS3 }
func (s *s3Store) AccessMode() string    { return s.cfg.AccessMode }
func (s *s3Store) ChunkSize() int64      { return s.chunkSize }
func (s *s3Store) MaxObjectBytes() int64 { return s.maxObject }
func (s *s3Store) MaxParts() int         { return s.maxParts }

func (s *s3Store) Capabilities() Caps {
	return Caps{
		DirectUpload:  true,
		ServerPutPart: false,
		PresignGet:    s.cfg.AccessMode == AccessPrivate,
		PublicURL:     s.cfg.AccessMode == AccessPublic,
	}
}

func (s *s3Store) Initiate(ctx context.Context, key string, opts InitiateOpts) (string, error) {
	key, err := NormalizeKey(key)
	if err != nil {
		return "", err
	}
	maxSize := opts.MaxSize
	if maxSize <= 0 {
		maxSize = s.maxObject
	}
	if maxSize > s.maxObject {
		return "", ErrTooLarge
	}
	in := &s3.CreateMultipartUploadInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}
	if opts.ContentType != "" {
		in.ContentType = aws.String(opts.ContentType)
	}
	out, err := s.client.CreateMultipartUpload(ctx, in)
	if err != nil {
		return "", fmt.Errorf("s3 create multipart: %w", err)
	}
	if out.UploadId == nil || *out.UploadId == "" {
		return "", ErrUploadNotFound
	}
	return *out.UploadId, nil
}

// PreparePart 下发单 part、短 TTL 的 UploadPart 预签名 URL。
func (s *s3Store) PreparePart(ctx context.Context, key, uploadID string, partNumber int, opts PartOpts) (PartPrep, error) {
	key, err := NormalizeKey(key)
	if err != nil {
		return PartPrep{}, err
	}
	if strings.TrimSpace(uploadID) == "" {
		return PartPrep{}, ErrUploadNotFound
	}
	if err := ValidatePartNumber(partNumber, s.maxParts); err != nil {
		return PartPrep{}, err
	}
	if opts.Size > 0 {
		if err := ValidatePartSize(opts.Size, s.chunkSize, s.maxObject); err != nil {
			return PartPrep{}, err
		}
	}

	in := &s3.UploadPartInput{
		Bucket:     aws.String(s.bucket),
		Key:        aws.String(key),
		UploadId:   aws.String(uploadID),
		PartNumber: aws.Int32(int32(partNumber)),
	}
	headers := map[string]string{}
	// 若调用方预声明分片哈希，写入 ChecksumSHA256 以便签名绑定
	if opts.ChunkSHA256 != "" {
		raw, err := DecodeSHA256Hex(opts.ChunkSHA256)
		if err != nil {
			return PartPrep{}, err
		}
		in.ChecksumSHA256 = aws.String(base64.StdEncoding.EncodeToString(raw))
		headers["x-amz-checksum-sha256"] = *in.ChecksumSHA256
	}

	out, err := s.presign.PresignUploadPart(ctx, in, s3.WithPresignExpires(s.presignTTL))
	if err != nil {
		return PartPrep{}, fmt.Errorf("s3 presign upload part: %w", err)
	}
	return PartPrep{
		UploadURL: out.URL,
		Headers:   headers,
		ExpiresAt: time.Now().UTC().Add(s.presignTTL),
	}, nil
}

func (s *s3Store) PutPart(ctx context.Context, key, uploadID string, partNumber int, r io.Reader, size int64, chunkSHA256 string) (PartResult, error) {
	return PartResult{}, ErrDirectUploadOnly
}

func (s *s3Store) ListParts(ctx context.Context, key, uploadID string) ([]PartResult, error) {
	key, err := NormalizeKey(key)
	if err != nil {
		return nil, err
	}
	var (
		out     []PartResult
		token   *string
	)
	for {
		resp, err := s.client.ListParts(ctx, &s3.ListPartsInput{
			Bucket:           aws.String(s.bucket),
			Key:              aws.String(key),
			UploadId:         aws.String(uploadID),
			PartNumberMarker: token,
		})
		if err != nil {
			return nil, fmt.Errorf("s3 list parts: %w", err)
		}
		for _, p := range resp.Parts {
			item := PartResult{
				PartNumber: int(aws.ToInt32(p.PartNumber)),
				Size:       aws.ToInt64(p.Size),
				ETag:       strings.Trim(aws.ToString(p.ETag), `"`),
			}
			if p.ChecksumSHA256 != nil {
				// S3 返回 base64，转 hex 便于与本地模型对齐
				if raw, err := base64.StdEncoding.DecodeString(*p.ChecksumSHA256); err == nil && len(raw) == 32 {
					item.SHA256 = HexSHA256(raw)
				}
			}
			out = append(out, item)
		}
		if !aws.ToBool(resp.IsTruncated) {
			break
		}
		token = resp.NextPartNumberMarker
	}
	return out, nil
}

// Complete 用服务端可信分片列表算 Merkle 后完成 multipart。
func (s *s3Store) Complete(ctx context.Context, key, uploadID string, parts []PartResult, expectMerkle string) (ObjectInfo, error) {
	key, err := NormalizeKey(key)
	if err != nil {
		return ObjectInfo{}, err
	}
	if len(parts) == 0 {
		return ObjectInfo{}, ErrInvalidPart
	}
	// 1. 规范化 part 顺序与大小
	ordered := append([]PartResult(nil), parts...)
	sortPartResults(ordered)
	digests := make([]string, 0, len(ordered))
	var total int64
	completed := make([]types.CompletedPart, 0, len(ordered))
	for i, p := range ordered {
		if p.PartNumber != i+1 {
			return ObjectInfo{}, fmt.Errorf("%w: missing part %d", ErrInvalidPart, i+1)
		}
		if i < len(ordered)-1 && p.Size != s.chunkSize && p.Size != 0 {
			// Size 可能由客户端上报；为 0 时跳过非最后片硬校验（依赖 S3）
			return ObjectInfo{}, fmt.Errorf("%w: non-final part size", ErrInvalidPart)
		}
		if p.ETag == "" {
			return ObjectInfo{}, ErrInvalidPart
		}
		total += p.Size
		cp := types.CompletedPart{
			ETag:       aws.String(p.ETag),
			PartNumber: aws.Int32(int32(p.PartNumber)),
		}
		if p.SHA256 != "" {
			raw, err := DecodeSHA256Hex(p.SHA256)
			if err != nil {
				return ObjectInfo{}, err
			}
			digests = append(digests, p.SHA256)
			cp.ChecksumSHA256 = aws.String(base64.StdEncoding.EncodeToString(raw))
		}
		completed = append(completed, cp)
	}
	if total > s.maxObject && total > 0 {
		return ObjectInfo{}, ErrTooLarge
	}

	var integ Integrity
	if len(digests) == len(ordered) {
		integ, err = BuildIntegrity(s.chunkSize, digests, "")
		if err != nil {
			return ObjectInfo{}, err
		}
		if expectMerkle != "" && !EqualSHA256Hex(expectMerkle, integ.MerkleSHA256) {
			return ObjectInfo{}, ErrMerkleMismatch
		}
	} else if expectMerkle != "" {
		return ObjectInfo{}, fmt.Errorf("%w: missing part digests for merkle", ErrChecksumMismatch)
	}

	// 2. CompleteMultipartUpload
	_, err = s.client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(s.bucket),
		Key:      aws.String(key),
		UploadId: aws.String(uploadID),
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: completed,
		},
	})
	if err != nil {
		return ObjectInfo{}, fmt.Errorf("s3 complete multipart: %w", err)
	}
	return ObjectInfo{Key: key, Size: total, Integrity: integ}, nil
}

func (s *s3Store) Abort(ctx context.Context, key, uploadID string) error {
	key, err := NormalizeKey(key)
	if err != nil {
		return err
	}
	_, err = s.client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(s.bucket),
		Key:      aws.String(key),
		UploadId: aws.String(uploadID),
	})
	return err
}

func (s *s3Store) Head(ctx context.Context, key string) (ObjectInfo, error) {
	key, err := NormalizeKey(key)
	if err != nil {
		return ObjectInfo{}, err
	}
	out, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return ObjectInfo{}, ErrObjectNotFound
	}
	return ObjectInfo{Key: key, Size: aws.ToInt64(out.ContentLength)}, nil
}

func (s *s3Store) Open(ctx context.Context, key string) (io.ReadCloser, ObjectInfo, error) {
	key, err := NormalizeKey(key)
	if err != nil {
		return nil, ObjectInfo{}, err
	}
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, ObjectInfo{}, ErrObjectNotFound
	}
	info := ObjectInfo{Key: key, Size: aws.ToInt64(out.ContentLength)}
	return out.Body, info, nil
}

func (s *s3Store) PresignGet(ctx context.Context, key string, expire time.Duration) (string, error) {
	if s.cfg.AccessMode != AccessPrivate {
		return "", ErrNotSupported
	}
	key, err := NormalizeKey(key)
	if err != nil {
		return "", err
	}
	if expire <= 0 {
		expire = s.presignTTL
	}
	out, err := s.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expire))
	if err != nil {
		return "", fmt.Errorf("s3 presign get: %w", err)
	}
	return out.URL, nil
}

func (s *s3Store) PublicURL(key string) (string, error) {
	if s.cfg.AccessMode != AccessPublic {
		return "", ErrNotSupported
	}
	return BuildS3PublicURL(s.cfg.PublicBaseURL, key)
}

func (s *s3Store) Delete(ctx context.Context, key string) error {
	key, err := NormalizeKey(key)
	if err != nil {
		return err
	}
	_, err = s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

func (s *s3Store) GCExpired(ctx context.Context, now time.Time) (int, error) {
	// 列出未完成 multipart 并按发起时间清理；失败不致命
	var (
		removed int
		keyMarker, uploadMarker *string
	)
	expireBefore := now.UTC().Add(-time.Duration(s.cfg.UploadExpireSeconds) * time.Second)
	for {
		out, err := s.client.ListMultipartUploads(ctx, &s3.ListMultipartUploadsInput{
			Bucket:         aws.String(s.bucket),
			KeyMarker:      keyMarker,
			UploadIdMarker: uploadMarker,
		})
		if err != nil {
			return removed, nil
		}
		for _, u := range out.Uploads {
			if u.Initiated != nil && u.Initiated.Before(expireBefore) {
				_, _ = s.client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{
					Bucket:   aws.String(s.bucket),
					Key:      u.Key,
					UploadId: u.UploadId,
				})
				removed++
			}
		}
		if !aws.ToBool(out.IsTruncated) {
			break
		}
		keyMarker = out.NextKeyMarker
		uploadMarker = out.NextUploadIdMarker
	}
	return removed, nil
}

func sortPartResults(parts []PartResult) {
	for i := 0; i < len(parts); i++ {
		for j := i + 1; j < len(parts); j++ {
			if parts[j].PartNumber < parts[i].PartNumber {
				parts[i], parts[j] = parts[j], parts[i]
			}
		}
	}
}

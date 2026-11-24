package captcha

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/big"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

const keyPrefix = "captcha:"

var digits = []byte("23456789ABCDEFGHJKLMNPQRSTUVWXYZ")

// TTLProvider 按请求取验证码 TTL。
type TTLProvider func(ctx context.Context) time.Duration

// LengthProvider 按请求取验证码位数。
type LengthProvider func(ctx context.Context) int

// Store 基于 Redis 的一次性图形验证码。
type Store struct {
	rdb      *redis.Redis
	ttlFn    TTLProvider
	lengthFn LengthProvider
}

func NewStore(rdb *redis.Redis, ttlFn TTLProvider, lengthFn LengthProvider) *Store {
	if ttlFn == nil {
		ttlFn = func(context.Context) time.Duration { return 120 * time.Second }
	}
	if lengthFn == nil {
		lengthFn = func(context.Context) int { return 4 }
	}
	return &Store{rdb: rdb, ttlFn: ttlFn, lengthFn: lengthFn}
}

// Issue 生成验证码，返回 id 与 PNG base64（不含 data URI 前缀）。
func (s *Store) Issue(ctx context.Context) (id, b64 string, err error) {
	// 1. 取当前长度/TTL（动态配置），生成答案与 id
	length := s.lengthFn(ctx)
	if length <= 0 {
		length = 4
	}
	code, err := randomCode(length)
	if err != nil {
		return "", "", err
	}
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	id = hex.EncodeToString(buf)
	// 2. 画图并编码
	img := render(code)
	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		return "", "", err
	}
	b64 = base64.StdEncoding.EncodeToString(out.Bytes())
	// 3. 写入 Redis（一次性）
	ttl := s.ttlFn(ctx)
	if ttl <= 0 {
		ttl = 120 * time.Second
	}
	if err := s.rdb.SetexCtx(ctx, keyPrefix+id, code, int(ttl.Seconds())); err != nil {
		return "", "", err
	}
	return id, b64, nil
}

// Verify 校验并删除；大小写不敏感。
func (s *Store) Verify(ctx context.Context, id, code string) bool {
	if id == "" || code == "" {
		return false
	}
	key := keyPrefix + id
	val, err := s.rdb.GetCtx(ctx, key)
	_, _ = s.rdb.DelCtx(ctx, key)
	if err != nil || val == "" {
		return false
	}
	return strings.EqualFold(val, strings.TrimSpace(code))
}

func randomCode(n int) (string, error) {
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		b[i] = digits[idx.Int64()]
	}
	return string(b), nil
}

// render 简易点阵数字/字母图，避免引入第三方验证码库。
func render(code string) image.Image {
	const cellW, cellH = 28, 40
	w := cellW*len(code) + 20
	h := cellH + 20
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.RGBA{R: 245, G: 247, B: 250, A: 255}}, image.Point{}, draw.Src)
	// 噪点
	for i := 0; i < w*h/8; i++ {
		x, _ := rand.Int(rand.Reader, big.NewInt(int64(w)))
		y, _ := rand.Int(rand.Reader, big.NewInt(int64(h)))
		img.Set(int(x.Int64()), int(y.Int64()), color.RGBA{R: 180, G: 180, B: 180, A: 255})
	}
	for i, ch := range code {
		drawGlyph(img, 10+i*cellW, 10, ch)
	}
	return img
}

func drawGlyph(img *image.RGBA, x, y int, ch rune) {
	// 用 5x7 简易点阵近似，足够局域网登录防机器人
	patterns := map[rune][]string{
		'2': {"11110", "00001", "00010", "00100", "01000", "10000", "11111"},
		'3': {"11110", "00001", "00001", "01110", "00001", "00001", "11110"},
		'4': {"00010", "00110", "01010", "10010", "11111", "00010", "00010"},
		'5': {"11111", "10000", "11110", "00001", "00001", "00001", "11110"},
		'6': {"01110", "10000", "10000", "11110", "10001", "10001", "01110"},
		'7': {"11111", "00001", "00010", "00100", "01000", "01000", "01000"},
		'8': {"01110", "10001", "10001", "01110", "10001", "10001", "01110"},
		'9': {"01110", "10001", "10001", "01111", "00001", "00001", "01110"},
		'A': {"01110", "10001", "10001", "11111", "10001", "10001", "10001"},
		'B': {"11110", "10001", "10001", "11110", "10001", "10001", "11110"},
		'C': {"01110", "10001", "10000", "10000", "10000", "10001", "01110"},
		'D': {"11110", "10001", "10001", "10001", "10001", "10001", "11110"},
		'E': {"11111", "10000", "10000", "11110", "10000", "10000", "11111"},
		'F': {"11111", "10000", "10000", "11110", "10000", "10000", "10000"},
		'G': {"01110", "10001", "10000", "10111", "10001", "10001", "01110"},
		'H': {"10001", "10001", "10001", "11111", "10001", "10001", "10001"},
		'J': {"00001", "00001", "00001", "00001", "10001", "10001", "01110"},
		'K': {"10001", "10010", "10100", "11000", "10100", "10010", "10001"},
		'L': {"10000", "10000", "10000", "10000", "10000", "10000", "11111"},
		'M': {"10001", "11011", "10101", "10001", "10001", "10001", "10001"},
		'N': {"10001", "11001", "10101", "10011", "10001", "10001", "10001"},
		'P': {"11110", "10001", "10001", "11110", "10000", "10000", "10000"},
		'Q': {"01110", "10001", "10001", "10001", "10101", "10010", "01101"},
		'R': {"11110", "10001", "10001", "11110", "10100", "10010", "10001"},
		'S': {"01111", "10000", "10000", "01110", "00001", "00001", "11110"},
		'T': {"11111", "00100", "00100", "00100", "00100", "00100", "00100"},
		'U': {"10001", "10001", "10001", "10001", "10001", "10001", "01110"},
		'V': {"10001", "10001", "10001", "10001", "10001", "01010", "00100"},
		'W': {"10001", "10001", "10001", "10001", "10101", "11011", "10001"},
		'X': {"10001", "10001", "01010", "00100", "01010", "10001", "10001"},
		'Y': {"10001", "10001", "01010", "00100", "00100", "00100", "00100"},
		'Z': {"11111", "00001", "00010", "00100", "01000", "10000", "11111"},
	}
	pat, ok := patterns[ch]
	if !ok {
		pat = patterns['X']
	}
	fg := color.RGBA{R: 30, G: 60, B: 120, A: 255}
	scale := 4
	for row, line := range pat {
		for col, c := range line {
			if c == '1' {
				for dy := 0; dy < scale; dy++ {
					for dx := 0; dx < scale; dx++ {
						img.Set(x+col*scale+dx, y+row*scale+dy, fg)
					}
				}
			}
		}
	}
}

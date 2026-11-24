package pwdcipher

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"math/big"
	"strings"

	"github.com/tjfoc/gmsm/sm2"
)

// Cipher 密码传输加解密（实现细节不对外暴露）。
type Cipher struct {
	priv *sm2.PrivateKey
}

// NewFromPrivateKeyHex 从私钥十六进制加载；失败返回 error。
func NewFromPrivateKeyHex(privHex string) (*Cipher, error) {
	privHex = strings.TrimSpace(privHex)
	if privHex == "" {
		return nil, errors.New("private key required")
	}
	raw, err := hex.DecodeString(privHex)
	if err != nil {
		return nil, errors.New("invalid private key")
	}
	curve := sm2.P256Sm2()
	d := new(big.Int).SetBytes(raw)
	priv := new(sm2.PrivateKey)
	priv.PublicKey.Curve = curve
	priv.D = d
	priv.PublicKey.X, priv.PublicKey.Y = curve.ScalarBaseMult(d.Bytes())
	if !curve.IsOnCurve(priv.PublicKey.X, priv.PublicKey.Y) {
		return nil, errors.New("invalid private key")
	}
	return &Cipher{priv: priv}, nil
}

// PublicKeyHex 返回公钥十六进制（未压缩点）。
func (c *Cipher) PublicKeyHex() string {
	if c == nil || c.priv == nil {
		return ""
	}
	x := c.priv.PublicKey.X.Bytes()
	y := c.priv.PublicKey.Y.Bytes()
	xb := pad32(x)
	yb := pad32(y)
	out := make([]byte, 1+64)
	out[0] = 0x04
	copy(out[1:33], xb)
	copy(out[33:], yb)
	return hex.EncodeToString(out)
}

// Decrypt 解密密码密文（hex）；失败不区分原因。
func (c *Cipher) Decrypt(cipherHex string) (string, error) {
	if c == nil || c.priv == nil {
		return "", errors.New("cipher unavailable")
	}
	cipherHex = strings.TrimSpace(cipherHex)
	if cipherHex == "" {
		return "", errors.New("empty cipher")
	}
	raw, err := hex.DecodeString(cipherHex)
	if err != nil {
		return "", errors.New("invalid cipher")
	}
	plain, err := sm2.Decrypt(c.priv, raw, sm2.C1C3C2)
	if err != nil {
		return "", errors.New("invalid cipher")
	}
	return string(plain), nil
}

// Encrypt 仅供测试/冒烟使用。
func (c *Cipher) Encrypt(plain string) (string, error) {
	if c == nil || c.priv == nil {
		return "", errors.New("cipher unavailable")
	}
	raw, err := sm2.Encrypt(&c.priv.PublicKey, []byte(plain), rand.Reader, sm2.C1C3C2)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}

func pad32(b []byte) []byte {
	if len(b) >= 32 {
		return b[len(b)-32:]
	}
	out := make([]byte, 32)
	copy(out[32-len(b):], b)
	return out
}

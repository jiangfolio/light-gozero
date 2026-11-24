package idgen

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const (
	ModeUUIDV7        = "uuidv7"
	ModeAutoIncrement = "auto_increment"
)

// Generator 统一 ID 生成门面；业务层只消费 string。
type Generator interface {
	Next() (string, error)
	Kind() string
}

// New 按配置模式装配生成器。
func New(mode string) (Generator, error) {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", ModeUUIDV7:
		return uuidV7Generator{}, nil
	case ModeAutoIncrement:
		return autoIncrGenerator{}, nil
	default:
		return nil, fmt.Errorf("unsupported IdGen.Mode %q", mode)
	}
}

type uuidV7Generator struct{}

func (uuidV7Generator) Kind() string { return ModeUUIDV7 }

func (uuidV7Generator) Next() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

type autoIncrGenerator struct{}

func (autoIncrGenerator) Kind() string { return ModeAutoIncrement }

// Next 自增模式交给数据库，应用侧返回空串。
func (autoIncrGenerator) Next() (string, error) { return "", nil }

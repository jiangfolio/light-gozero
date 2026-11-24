// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"github.com/jiangbyte/light-gozero/common/storage"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	Storage storage.Config
}

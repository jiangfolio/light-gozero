// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"fmt"

	"github.com/jiangbyte/light-gozero/common/storage"
	"github.com/jiangbyte/light-gozero/service/demoservice/api/internal/config"
)

type ServiceContext struct {
	Config    config.Config
	FileStore storage.Store
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 1. 按 yaml 构造对象存储；失败直接退出，避免半残服务
	store, err := storage.New(c.Storage)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Storage.Provider=%s AccessMode=%s ChunkSize=%d\n",
		store.Provider(), store.AccessMode(), store.ChunkSize())

	return &ServiceContext{
		Config:    c,
		FileStore: store,
	}
}

package model

import (
	"context"

	"gorm.io/gorm"

	"github.com/jiangbyte/light-gozero/service/system/api/internal/model/audit"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/model/config"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/model/dict"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/model/org"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/model/resource"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/model/role"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/model/shortcut"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/model/user"
)

// Repo 聚合各域数据访问入口（目录对齐 desc）。
type Repo struct {
	db       *gorm.DB
	Org      *org.Repo
	User     *user.Repo
	Role     *role.Repo
	Resource *resource.Repo
	Audit    *audit.Repo
	Config   *config.Repo
	Dict     *dict.Repo
	Shortcut *shortcut.Repo
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{
		db:       db,
		Org:      org.NewRepo(db),
		User:     user.NewRepo(db),
		Role:     role.NewRepo(db),
		Resource: resource.NewRepo(db),
		Audit:    audit.NewRepo(db),
		Config:   config.NewRepo(db),
		Dict:     dict.NewRepo(db),
		Shortcut: shortcut.NewRepo(db),
	}
}

// Transaction 在短事务中执行写操作；回调内使用 txRepo，禁止远程长 IO。
func (r *Repo) Transaction(ctx context.Context, fn func(txRepo *Repo) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(NewRepo(tx))
	})
}

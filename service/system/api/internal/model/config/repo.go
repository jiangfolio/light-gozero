package config

import (
	"context"

	"gorm.io/gorm"

	"github.com/jiangbyte/light-gozero/common/appcfg"
	"github.com/jiangbyte/light-gozero/common/batchutil"
)

// Repo 系统配置数据访问。
type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo { return &Repo{db: db} }

func (r *Repo) withCtx(ctx context.Context) *gorm.DB { return r.db.WithContext(ctx) }

// ListEnabled 列出启用配置，供 AppConfig 缓存装载。
func (r *Repo) ListEnabled(ctx context.Context) ([]appcfg.Item, error) {
	var rows []SysConfig
	err := r.withCtx(ctx).
		Select("config_key", "config_value", "value_type", "group_name", "status").
		Where("status = ?", 1).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]appcfg.Item, 0, len(rows))
	for i := range rows {
		out = append(out, appcfg.Item{
			ConfigKey:   rows[i].ConfigKey,
			ConfigValue: rows[i].ConfigValue,
			ValueType:   rows[i].ValueType,
			GroupName:   rows[i].GroupName,
			Status:      rows[i].Status,
		})
	}
	return out, nil
}

func (r *Repo) FindById(ctx context.Context, id string) (*SysConfig, error) {
	var row SysConfig
	err := r.withCtx(ctx).Where("id = ?", id).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *Repo) FindByKey(ctx context.Context, key string) (*SysConfig, error) {
	var row SysConfig
	err := r.withCtx(ctx).Where("config_key = ?", key).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// ListByKeys 按配置键批量加载。
func (r *Repo) ListByKeys(ctx context.Context, keys []string) ([]SysConfig, error) {
	keys = batchutil.UniqueStrings(keys)
	if len(keys) == 0 {
		return nil, nil
	}
	var list []SysConfig
	for _, part := range batchutil.Chunk(keys, batchutil.InSize) {
		var chunk []SysConfig
		if err := r.withCtx(ctx).Where("config_key IN ?", part).Find(&chunk).Error; err != nil {
			return nil, err
		}
		list = append(list, chunk...)
	}
	return list, nil
}

func (r *Repo) Page(ctx context.Context, groupName, configKey string, status int64, hasStatus bool, offset, limit int) ([]SysConfig, int64, error) {
	q := r.withCtx(ctx).Model(&SysConfig{})
	if groupName != "" {
		q = q.Where("group_name = ?", groupName)
	}
	if configKey != "" {
		q = q.Where("config_key LIKE ?", "%"+configKey+"%")
	}
	if hasStatus {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []SysConfig
	err := q.Order("group_name asc, config_key asc").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *Repo) Save(ctx context.Context, row *SysConfig) error {
	return r.withCtx(ctx).Where("id = ?", row.Id).Save(row).Error
}

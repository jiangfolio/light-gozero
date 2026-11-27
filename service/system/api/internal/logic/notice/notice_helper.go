package notice

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jiangbyte/light-gozero/common/convert"
	"github.com/jiangbyte/light-gozero/common/errorx"
	noticemodel "github.com/jiangbyte/light-gozero/service/system/api/internal/model/notice"
	"github.com/jiangbyte/light-gozero/service/system/api/internal/types"
)

const (
	maxTitleRunes   = 200
	maxRemarkRunes  = 255
	maxContentRunes = 20000
)

func toNoticeInfo(row *noticemodel.SysNotice) types.NoticeInfo {
	publishAt := ""
	if row.PublishAt != nil {
		publishAt = convert.TimeToString(*row.PublishAt)
	}
	return types.NoticeInfo{
		Id:        row.Id,
		Title:     row.Title,
		Content:   row.Content,
		Type:      convert.Int64ToString(row.Type),
		Status:    convert.Int64ToString(row.Status),
		Sort:      convert.Int64ToString(row.Sort),
		PublishAt: publishAt,
		Remark:    row.Remark,
		CreatedBy: row.CreatedBy,
		UpdatedBy: row.UpdatedBy,
		CreatedAt: convert.TimeToString(row.CreatedAt),
		UpdatedAt: convert.TimeToString(row.UpdatedAt),
	}
}

/** 登录可读视图：不暴露操作人内部 ID。 */
func toPublishedNoticeInfo(row *noticemodel.SysNotice) types.NoticeInfo {
	info := toNoticeInfo(row)
	info.CreatedBy = ""
	info.UpdatedBy = ""
	return info
}

/** 解析公告类型：1 通知 / 2 公告。 */
func parseNoticeType(s string, def int64) (int64, error) {
	if s == "" {
		return def, nil
	}
	v, err := convert.Int64FromString(s)
	if err != nil || (v != noticemodel.TypeNotice && v != noticemodel.TypeAnnounce) {
		return 0, errorx.New(400, "类型无效")
	}
	return v, nil
}

/** 解析公告状态：0 草稿 / 1 已发布。 */
func parseNoticeStatus(s string, def int64) (int64, error) {
	if s == "" {
		return def, nil
	}
	v, err := convert.Int64FromString(s)
	if err != nil || (v != noticemodel.StatusDraft && v != noticemodel.StatusPublished) {
		return 0, errorx.New(400, "状态无效")
	}
	return v, nil
}

/** 发布态补齐 publish_at；退回草稿则清空。 */
func applyPublishAt(row *noticemodel.SysNotice, now time.Time) {
	if row.Status == noticemodel.StatusPublished {
		if row.PublishAt == nil || row.PublishAt.IsZero() {
			t := now
			row.PublishAt = &t
		}
		return
	}
	row.PublishAt = nil
}

/** 清洗并校验标题/正文/备注长度，避免落库报内部错误。 */
func normalizeNoticeText(title, content, remark string, requireTitleContent bool) (string, string, string, error) {
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)
	remark = strings.TrimSpace(remark)
	if requireTitleContent {
		if title == "" || content == "" {
			return "", "", "", errorx.New(400, "标题和正文不能为空")
		}
	}
	if utf8.RuneCountInString(title) > maxTitleRunes {
		return "", "", "", errorx.New(400, "标题过长")
	}
	if utf8.RuneCountInString(content) > maxContentRunes {
		return "", "", "", errorx.New(400, "正文过长")
	}
	if utf8.RuneCountInString(remark) > maxRemarkRunes {
		return "", "", "", errorx.New(400, "备注过长")
	}
	return title, content, remark, nil
}

/** 操作摘要截断标题，避免日志过长。 */
func summaryTitle(title string) string {
	title = strings.TrimSpace(title)
	r := []rune(title)
	if len(r) > 40 {
		return string(r[:40]) + "…"
	}
	return title
}

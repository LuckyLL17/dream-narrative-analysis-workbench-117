package service

import (
	"fmt"
	"strings"
	"time"

	"dream117/internal/domain"
	"dream117/internal/store"
	"dream117/pkg/jsonutil"
)

type ExportService struct{ store *store.Store }

func NewExportService(
	data *store.Store,
) *ExportService {
	return &ExportService{store: data}
}
func (
	s *ExportService,
) JSON(
	userID string,
) ([]byte, error) {
	bundle := s.bundle(userID, "dream-workbench-json-v2")
	return jsonutil.MarshalIndent(bundle, "", "  ")
}
func (
	s *ExportService,
) Markdown(
	userID string,
) string {
	bundle := s.bundle(userID, "dream-workbench-markdown-v2")
	items := bundle.Dreams
	var builder strings.Builder
	builder.WriteString("# 我的梦境记录\n\n")
	builder.WriteString(fmt.Sprintf("导出时间：%s\n\n", bundle.ExportedAt.Format("2006-01-02 15:04:05 UTC")))
	if bundle.User.DisplayName != "" {
		builder.WriteString(fmt.Sprintf("记录者：%s\n\n", bundle.User.DisplayName))
	}
	if len(items) == 0 {
		builder.WriteString("暂时没有记录。\n")
		return builder.String()
	}
	for i := range items {
		d := items[i]
		builder.WriteString(fmt.Sprintf("## %s · %s\n\n", d.Title, d.DreamDate.Format("2006-01-02")))
		builder.WriteString(fmt.Sprintf("- 醒来情绪：%s\n- 清晰度：%d/10\n- 睡眠时长：%.1f 小时\n- 细节记忆：%s\n\n", d.Emotion, d.Clarity, d.SleepHours, detailLabel(d.RememberDetail)))
		builder.WriteString(d.Content + "\n\n")
		if len(d.Tags) > 0 {
			builder.WriteString("元素：" + strings.Join(d.Tags, "、") + "\n\n")
		}
		if len(d.Themes) > 0 {
			names := make(
				[]string,
				0,
				len(d.Themes),
			)
			for i := range d.Themes {
				theme := d.Themes[i]
				names = append(names,
					theme.Name)
			}
			builder.WriteString("主题：" + strings.Join(names, "、") + "\n\n")
		}
		builder.WriteString("---\n\n")
	}
	if len(bundle.Reports) > 0 {
		builder.WriteString("# 周报快照\n\n")
		for i := range bundle.Reports {
			report := bundle.Reports[i]
			builder.WriteString(fmt.Sprintf("- %s 至 %s：%d 条，平均睡眠 %.1f 小时，平均清晰度 %.1f 分\n", report.WeekStart.Format("2006-01-02"), report.WeekEnd.Format("2006-01-02"), report.TotalDreams, report.AvgSleep, report.AvgClarity))
		}
	}
	return builder.String()
}

func (
	s *ExportService,
) bundle(
	userID,
	format string,
) domain.ExportBundle {
	user, _ := s.store.FindUser(
		userID)
	user.PasswordHash = ""
	return domain.ExportBundle{ExportedAt: time.Now().UTC(), User: user, Dreams: s.store.AllDreams(userID), Elements: s.store.ListElements(userID), Reports: s.store.ListReports(userID), Format: format}
}
func detailLabel(
	value bool,
) string {
	if value {
		return "记得具体细节"
	}
	return "只有片段印象"
}

var _ domain.Dream

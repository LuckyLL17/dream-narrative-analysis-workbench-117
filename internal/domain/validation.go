package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidInput = errors.New(
		"输入不符合领域规则")
	ErrNotFound = errors.New(
		"记录不存在")
	ErrConflict = errors.New(
		"记录状态冲突")
	ErrUnauthorized = errors.New(
		"未授权")
)

func ValidateUser(
	email,
	name,
	password string,
) error {
	if !strings.Contains(email, "@") || len([]rune(email)) > 120 {
		return fmt.Errorf("%w: 邮箱格式不正确", ErrInvalidInput)
	}
	if len([]rune(strings.TrimSpace(name))) < 2 {
		return fmt.Errorf("%w: 昵称至少需要两个字符", ErrInvalidInput)
	}
	if len(password) < 8 || len(password) > 72 {
		return fmt.Errorf("%w: 密码长度需要在 8 到 72 个字符之间", ErrInvalidInput)
	}
	return nil
}

func ValidateDream(
	d Dream,
) error {
	if strings.TrimSpace(d.Title) == "" || len([]rune(d.Title)) > 80 {
		return fmt.Errorf("%w: 标题不能为空且不能超过 80 个字符", ErrInvalidInput)
	}
	if len([]rune(strings.TrimSpace(d.Content))) < 8 {
		return fmt.Errorf("%w: 梦境描述至少需要 8 个字符", ErrInvalidInput)
	}
	if d.DreamDate.IsZero() || d.WakeTime.IsZero() {
		return fmt.Errorf("%w: 梦境日期和醒来时间不能为空", ErrInvalidInput)
	}
	if d.DreamDate.After(time.Now().Add(24 * time.Hour)) {
		return fmt.Errorf("%w: 梦境日期不能远在未来", ErrInvalidInput)
	}
	if d.SleepHours < 0 || d.SleepHours > 24 {
		return fmt.Errorf("%w: 睡眠时长需要在 0 到 24 小时之间", ErrInvalidInput)
	}
	if d.Clarity < 1 || d.Clarity > 10 {
		return fmt.Errorf("%w: 清晰度需要在 1 到 10 分之间", ErrInvalidInput)
	}
	if _, err := ParseEmotion(string(d.Emotion)); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}
	if len(d.Tags) > 24 {
		return fmt.Errorf("%w: 单条记录最多关联 24 个元素", ErrInvalidInput)
	}
	return nil
}

func ValidateWindow(
	from,
	to time.Time,
) error {
	if from.IsZero() || to.IsZero() || to.Before(from) {
		return fmt.Errorf("%w: 分析时间范围无效", ErrInvalidInput)
	}
	if to.Sub(from) > 366*24*time.Hour {
		return fmt.Errorf("%w: 单次分析不能超过 366 天", ErrInvalidInput)
	}
	return nil
}

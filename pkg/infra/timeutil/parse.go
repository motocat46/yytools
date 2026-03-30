// Package timeutil.

// 版权所有(Copyright)[yangyuan]
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// 作者:  yangyuan
// 创建日期:2026/3/26
package timeutil

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ── 公开 API ──────────────────────────────────────────────────────────────────

// Parse 解析日期或日期时间字符串，返回 time.Time（时区为 time.Local）。
//
// 支持格式：
//   - 纯日期：    "2024-01-05" / "2024/1/5"（月日可省略前导零，分隔符 - 或 /）
//   - 日期时间：  "2024-01-05 09:05:03" / "2024/1/5 9:5:3"（秒可省略，省略时默认 0）
//
// 输入前后空白和中间多余空白自动处理。不支持纯时间字符串（如 "09:05:03"）。
// 需要指定时区时使用 ParseInLoc。
func Parse(s string) (time.Time, error) {
	return ParseInLoc(s, time.Local)
}

// ParseInLoc 解析同 Parse，以 loc 指定的时区解释日期时间字符串。
// 用于服务器时区与业务时区不一致的场景（如服务器在 UTC+0，业务使用 Asia/Shanghai）。
//
// 示例（loc = Asia/Shanghai）：
//
//	"2024-03-15 10:00:00" → 2024-03-15 10:00:00 CST（= UTC+8 的 02:00:00 UTC）
func ParseInLoc(s string, loc *time.Location) (time.Time, error) {
	orig := s
	s = strings.TrimSpace(s)
	if s == "" {
		if orig != "" {
			return time.Time{}, fmt.Errorf("timeutil.Parse: 输入为空（原始输入: %q）", orig)
		}
		return time.Time{}, fmt.Errorf("timeutil.Parse: 输入为空")
	}
	fields := strings.Fields(s)
	switch len(fields) {
	case 1:
		return parseDate(fields[0], loc)
	case 2:
		return parseDateTime(fields[0], fields[1], loc)
	default:
		return time.Time{}, fmt.Errorf("timeutil.Parse: 格式不支持（原始输入: %q）", s)
	}
}

// ParseUnixMilli 解析同 Parse，成功时返回 Unix 毫秒时间戳（时区为 time.Local）。
// 非法输入返回 0 和非 nil 错误。
func ParseUnixMilli(s string) (int64, error) {
	return ParseUnixMilliInLoc(s, time.Local)
}

// ParseUnixMilliInLoc 解析同 ParseInLoc，成功时返回 Unix 毫秒时间戳。
// 语义见 ParseInLoc。
func ParseUnixMilliInLoc(s string, loc *time.Location) (int64, error) {
	t, err := ParseInLoc(s, loc)
	if err != nil {
		return 0, err
	}
	return t.UnixMilli(), nil
}

// ── 内部实现 ──────────────────────────────────────────────────────────────────

func parseDate(s string, loc *time.Location) (time.Time, error) {
	normalized, err := normalizeDate(s)
	if err != nil {
		return time.Time{}, fmt.Errorf("timeutil.Parse: %w（原始输入: %q）", err, s)
	}
	t, err := time.ParseInLocation("2006-01-02", normalized, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("timeutil.Parse: 日期不合法（原始输入: %q）: %w", s, err)
	}
	return t, nil
}

func parseDateTime(datePart, timePart string, loc *time.Location) (time.Time, error) {
	orig := datePart + " " + timePart
	normDate, err := normalizeDate(datePart)
	if err != nil {
		return time.Time{}, fmt.Errorf("timeutil.Parse: %w（原始输入: %q）", err, orig)
	}
	normTime, err := normalizeTime(timePart)
	if err != nil {
		return time.Time{}, fmt.Errorf("timeutil.Parse: %w（原始输入: %q）", err, orig)
	}
	t, err := time.ParseInLocation("2006-01-02 15:04:05", normDate+" "+normTime, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("timeutil.Parse: 日期时间不合法（原始输入: %q %q）: %w", datePart, timePart, err)
	}
	return t, nil
}

// normalizeDate 将日期字符串规范化为 "2006-01-02" 格式。
// 支持 - 或 / 分隔符，月日可省略前导零，年份必须为 4 位。
// 不允许在同一字符串中混用 - 和 /。
func normalizeDate(s string) (string, error) {
	hasDash := strings.Contains(s, "-")
	hasSlash := strings.Contains(s, "/")
	if hasDash && hasSlash {
		return "", fmt.Errorf("日期分隔符混用（同时含 - 和 /）")
	}
	s = strings.ReplaceAll(s, "/", "-")
	parts := strings.Split(s, "-")
	if len(parts) != 3 {
		return "", fmt.Errorf("日期格式不合法，期望 3 段，实际 %d 段", len(parts))
	}
	year, month, day := parts[0], parts[1], parts[2]
	if len(year) != 4 {
		return "", fmt.Errorf("年份必须为 4 位")
	}
	if _, err := strconv.Atoi(year); err != nil {
		return "", fmt.Errorf("年份非数字")
	}
	month, err := padTwo(month)
	if err != nil {
		return "", err
	}
	day, err = padTwo(day)
	if err != nil {
		return "", err
	}
	return year + "-" + month + "-" + day, nil
}

// normalizeTime 将时间字符串规范化为 "15:04:05" 格式。
// 支持 HH:MM:SS 和 HH:MM（省略秒时补 00），各段可省略前导零，但不超过 2 位。
func normalizeTime(s string) (string, error) {
	parts := strings.Split(s, ":")
	if len(parts) < 2 || len(parts) > 3 {
		return "", fmt.Errorf("时间格式不合法，期望 2 或 3 段，实际 %d 段", len(parts))
	}
	if len(parts) == 2 {
		parts = append(parts, "0") // 省略秒时补 0
	}
	result := make([]string, 3)
	for i, p := range parts {
		padded, err := padTwo(p)
		if err != nil {
			return "", err
		}
		result[i] = padded
	}
	return result[0] + ":" + result[1] + ":" + result[2], nil
}

// padTwo 验证 s 为纯数字并补前导零至 2 位。
// 不包含原始输入上下文，由调用方负责在错误中添加。
func padTwo(s string) (string, error) {
	if s == "" {
		return "", fmt.Errorf("数字段为空")
	}
	if _, err := strconv.Atoi(s); err != nil {
		return "", fmt.Errorf("数字段非数字 %q", s)
	}
	if len(s) == 1 {
		return "0" + s, nil
	}
	if len(s) > 2 {
		return "", fmt.Errorf("数字段超过 2 位 %q", s)
	}
	return s, nil
}

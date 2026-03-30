package timeutil

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"
	"time"
)

func TestNormalizeTime(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"标准时分秒", "09:05:03", "09:05:03", false},
		{"省略前导零", "9:5:3", "09:05:03", false},
		{"省略秒", "09:05", "09:05:00", false},
		{"省略秒省略前导零", "9:5", "09:05:00", false},
		{"含空段", "09:05:", "", true},
		{"段数过多", "09:05:03:extra", "", true},
		{"段数不足", "09", "", true},
		{"非数字", "09:ab:03", "", true},
		{"小时超2位", "009:05:03", "", true},
		{"分钟超2位", "09:005:03", "", true},
		{"空字符串", "", "", true},
		{"秒超2位", "09:05:003", "", true},
		{"中间空段", "09::03", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := normalizeTime(tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("normalizeTime(%q) error = %v, wantErr %v", tc.input, err, tc.wantErr)
			}
			if !tc.wantErr && got != tc.want {
				t.Errorf("normalizeTime(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestNormalizeDate(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"标准格式", "2024-01-05", "2024-01-05", false},
		{"斜杠分隔", "2024/01/05", "2024-01-05", false},
		{"省略前导零", "2024-1-5", "2024-01-05", false},
		{"斜杠省略前导零", "2024/1/5", "2024-01-05", false},
		{"混用分隔符", "2024-01/05", "", true},
		{"年份非4位", "24-01-05", "", true},
		{"月份非数字", "2024-ab-05", "", true},
		{"段数不足", "2024-01", "", true},
		{"段数过多", "2024-01-05-extra", "", true},
		{"年份超4位", "20240-01-05", "", true},
		{"月日超2位", "2024-001-05", "", true},
		{"日超2位", "2024-01-005", "", true},
		{"空段连续分隔符", "2024--05", "", true},
		{"年份非数字", "20ab-01-05", "", true},
		{"日份非数字", "2024-01-ab", "", true},
		{"斜杠年份非4位", "24/01/05", "", true},
		{"空字符串", "", "", true},
		{"斜杠月份非数字", "2024/ab/05", "", true},
		{"斜杠月份超2位", "2024/001/05", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := normalizeDate(tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("normalizeDate(%q) error = %v, wantErr %v", tc.input, err, tc.wantErr)
			}
			if !tc.wantErr && got != tc.want {
				t.Errorf("normalizeDate(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestParse_InvalidInputs(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{"空字符串", ""},
		{"纯空白", "   "},
		{"纯时间字符串", "09:05:03"},
		{"日期段数不足", "2024-01"},
		{"日期段数过多", "2024-01-05-extra"},
		{"月份非数字", "2024-abc-05"},
		{"年份非4位", "24-01-05"},
		{"年份超4位", "20240-01-05"},
		{"混用分隔符", "2024-01/05"},
		{"月份超范围", "2024-13-01"},
		{"日期0", "2024-01-00"},
		{"月份0", "2024-00-05"},
		{"非闰年2月29日", "2023-02-29"},
		{"日期不合法", "2024-02-30"},
		{"小时超范围", "2024-01-05 25:00:00"},
		{"分钟超范围", "2024-01-05 09:60:00"},
		{"秒超范围", "2024-01-05 09:05:60"},
		{"时间含空段", "2024-01-05 09:05:"},
		{"时间段数过多", "2024-01-05 09:05:03:extra"},
		{"三个以上字段", "2024-01-05 09:05:03 extra"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.input)
			if err == nil {
				t.Errorf("Parse(%q) 期望错误，但得到 %v", tc.input, got)
				return
			}
			if !got.IsZero() {
				t.Errorf("Parse(%q) 错误时应返回零值，但得到 %v", tc.input, got)
			}
			// 错误信息应包含原始输入（空字符串输入跳过此检查）
			if tc.input != "" && !strings.Contains(err.Error(), tc.input) {
				t.Errorf("Parse(%q) 错误信息不含原始输入: %v", tc.input, err)
			}
		})
	}
}

func TestParse_Timezone(t *testing.T) {
	got, err := Parse("2024-01-05")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Location() != time.Local {
		t.Errorf("时区应为 time.Local，得到 %v", got.Location())
	}
}

func TestParse_ValidInputs(t *testing.T) {
	wantDate          := time.Date(2024, 3, 5, 0, 0, 0, 0, time.Local)
	wantDateTime      := time.Date(2024, 3, 5, 9, 5, 3, 0, time.Local)
	wantDateTimeNoSec := time.Date(2024, 3, 5, 9, 5, 0, 0, time.Local)

	cases := []struct {
		name  string
		input string
		want  time.Time
	}{
		// 纯日期
		{"标准日期", "2024-03-05", wantDate},
		{"斜杠日期", "2024/03/05", wantDate},
		{"省略前导零", "2024-3-5", wantDate},
		{"斜杠省略前导零", "2024/3/5", wantDate},
		// 日期时间
		{"标准日期时间", "2024-03-05 09:05:03", wantDateTime},
		{"省略前导零日期时间", "2024-3-5 9:5:3", wantDateTime},
		{"省略秒", "2024-03-05 09:05", wantDateTimeNoSec},
		{"斜杠日期时间", "2024/03/05 09:05:03", wantDateTime},
		// 空白处理
		{"前后空格", "  2024-03-05  ", wantDate},
		{"中间多空格", "2024-03-05  09:05:03", wantDateTime},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.input)
			if err != nil {
				t.Fatalf("Parse(%q) unexpected error: %v", tc.input, err)
			}
			if !got.Equal(tc.want) {
				t.Errorf("Parse(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestParseUnixMilli(t *testing.T) {
	// 合法输入：与 time.Date 直接构造的期望值对比，不依赖 Parse 的实现
	// 避免循环验证（ParseUnixMilli 内部调用 Parse，两者比较无法发现绝对值 bug）
	t.Run("合法输入绝对值验证", func(t *testing.T) {
		cases := []struct {
			input string
			want  int64
		}{
			{"2024-03-05 09:05:03", time.Date(2024, 3, 5, 9, 5, 3, 0, time.Local).UnixMilli()},
			{"2024-03-05", time.Date(2024, 3, 5, 0, 0, 0, 0, time.Local).UnixMilli()},
			{"2024/03/05", time.Date(2024, 3, 5, 0, 0, 0, 0, time.Local).UnixMilli()},
			{"2024-3-5", time.Date(2024, 3, 5, 0, 0, 0, 0, time.Local).UnixMilli()},
			{"2024-3-5 9:5:3", time.Date(2024, 3, 5, 9, 5, 3, 0, time.Local).UnixMilli()},
			{"2024-03-05 09:05", time.Date(2024, 3, 5, 9, 5, 0, 0, time.Local).UnixMilli()},
		}
		for _, tc := range cases {
			got, err := ParseUnixMilli(tc.input)
			if err != nil {
				t.Fatalf("ParseUnixMilli(%q) unexpected error: %v", tc.input, err)
			}
			if got != tc.want {
				t.Errorf("ParseUnixMilli(%q) = %d, want %d", tc.input, got, tc.want)
			}
		}
	})

	t.Run("非法输入返回0和err", func(t *testing.T) {
		cases := []string{
			"",
			"   ",
			"not-a-date",
			"09:05:03",
			"2024-13-01",
			"2024-01-05 25:00:00",
			"2024-01/05",
		}
		for _, input := range cases {
			got, err := ParseUnixMilli(input)
			if err == nil {
				t.Errorf("ParseUnixMilli(%q) 期望错误，但 err == nil", input)
				continue
			}
			if got != 0 {
				t.Errorf("ParseUnixMilli(%q) 非法输入应返回 0，得到 %d", input, got)
			}
		}
	})
}

func TestParse_CrossFormatConsistency(t *testing.T) {
	// 同一日期时间的所有格式变体，结果必须完全相同
	groups := [][]string{
		{"2024-01-05", "2024/01/05", "2024-1-5", "2024/1/5"},
		{"2024-01-05 09:05:03", "2024/01/05 09:05:03", "2024-1-5 9:5:3", "2024/1/5 9:5:3"},
		{"2024-01-05 09:05", "2024/01/05 09:05", "2024-1-5 9:5", "2024/1/5 9:5"},
	}
	for _, group := range groups {
		t.Run(group[0], func(t *testing.T) {
			ref, err := Parse(group[0])
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", group[0], err)
			}
			for _, variant := range group[1:] {
				got, err := Parse(variant)
				if err != nil {
					t.Fatalf("Parse(%q) error: %v", variant, err)
				}
				if !got.Equal(ref) {
					t.Errorf("格式不一致: Parse(%q)=%v != Parse(%q)=%v",
						variant, got, group[0], ref)
				}
			}
		})
	}
}

func TestParse_BoundaryValues(t *testing.T) {
	// 验证解析结果正确的用例（含具体日期断言）
	validCases := []struct {
		name  string
		input string
		want  time.Time
	}{
		{"1月1日", "2024-01-01", time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)},
		{"12月31日", "2024-12-31", time.Date(2024, 12, 31, 0, 0, 0, 0, time.Local)},
		{"时间最小值", "2024-01-05 0:0:0", time.Date(2024, 1, 5, 0, 0, 0, 0, time.Local)},
		{"时间最大值", "2024-01-05 23:59:59", time.Date(2024, 1, 5, 23, 59, 59, 0, time.Local)},
		{"闰年2月29日合法", "2024-02-29", time.Date(2024, 2, 29, 0, 0, 0, 0, time.Local)},
		{"1月31日合法", "2024-01-31", time.Date(2024, 1, 31, 0, 0, 0, 0, time.Local)},
		{"4月30日合法", "2024-04-30", time.Date(2024, 4, 30, 0, 0, 0, 0, time.Local)},
		{"2月28日合法（非闰年）", "2023-02-28", time.Date(2023, 2, 28, 0, 0, 0, 0, time.Local)},
	}
	for _, tc := range validCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.input)
			if err != nil {
				t.Fatalf("Parse(%q) unexpected error: %v", tc.input, err)
			}
			if !got.Equal(tc.want) {
				t.Errorf("Parse(%q) = %v, want %v", tc.input, got, tc.want)
			}
			if got.Location() != time.Local {
				t.Errorf("Parse(%q) location=%v, want time.Local", tc.input, got.Location())
			}
		})
	}

	// 验证应当报错的用例
	invalidCases := []struct {
		name  string
		input string
	}{
		{"非闰年2月29日非法", "2023-02-29"},
		{"4月31日非法", "2024-04-31"},
	}
	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.input)
			if err == nil {
				t.Errorf("Parse(%q) 期望错误，但得到 %v", tc.input, got)
			}
			if !got.IsZero() {
				t.Errorf("Parse(%q) 错误时应返回零值，但得到 %v", tc.input, got)
			}
		})
	}
}

func TestParse_Roundtrip(t *testing.T) {
	// 随机生成 time.Time → 格式化为标准字符串 → Parse → 与原始值比较（精度到秒）
	rng := rand.New(rand.NewPCG(42, 0))
	const n = 100_000
	for i := range n {
		year := 1970 + rng.IntN(130)
		month := time.Month(1 + rng.IntN(12))
		// 计算当月最后一天
		lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()
		day := 1 + rng.IntN(lastDay)
		hour := rng.IntN(24)
		minute := rng.IntN(60)
		second := rng.IntN(60)
		orig := time.Date(year, month, day, hour, minute, second, 0, time.Local)

		s := orig.Format("2006-01-02 15:04:05")
		got, err := Parse(s)
		if err != nil {
			t.Fatalf("第 %d 组 Parse(%q) error: %v", i, s, err)
		}
		if !got.Equal(orig) {
			t.Fatalf("第 %d 组 Roundtrip 失败: got %v, want %v（输入: %q）", i, got, orig, s)
		}
	}
}

func TestParse_RandomVariants(t *testing.T) {
	// 对同一日期生成所有格式变体，断言解析结果一致
	rng := rand.New(rand.NewPCG(99, 0))
	const n = 100_000
	for i := range n {
		year := 1970 + rng.IntN(130)
		month := time.Month(1 + rng.IntN(12))
		lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()
		day := 1 + rng.IntN(lastDay)
		ref := time.Date(year, month, day, 0, 0, 0, 0, time.Local)

		seps := []string{"-", "/"}
		for _, sep := range seps {
			full := fmt.Sprintf("%04d%s%02d%s%02d", year, sep, int(month), sep, day)
			short := fmt.Sprintf("%04d%s%d%s%d", year, sep, int(month), sep, day)

			for _, s := range []string{full, short} {
				got, err := Parse(s)
				if err != nil {
					t.Fatalf("第 %d 组 Parse(%q) error: %v", i, s, err)
				}
				if !got.Equal(ref) {
					t.Fatalf("第 %d 组 变体不一致: Parse(%q)=%v, want %v", i, s, got, ref)
				}
			}
		}
	}
}

func BenchmarkParse_Date(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, _ = Parse("2024-01-05")
	}
}

func BenchmarkParse_DateShort(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, _ = Parse("2024-1-5")
	}
}

func BenchmarkParse_DateTime(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, _ = Parse("2024-01-05 09:05:03")
	}
}

func BenchmarkParse_DateTimeShort(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, _ = Parse("2024-1-5 9:5")
	}
}

func BenchmarkParseUnixMilli(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, _ = ParseUnixMilli("2024-01-05 09:05:03")
	}
}

func TestParseInLoc(t *testing.T) {
	shanghai, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load Asia/Shanghai: %v", err)
	}
	utc := time.UTC

	// 同一字符串在不同时区解析出不同 Unix 时间戳
	const s = "2024-03-15 10:00:00"
	tShanghai, err := ParseInLoc(s, shanghai)
	if err != nil {
		t.Fatalf("ParseInLoc(%q, Shanghai): %v", s, err)
	}
	tUTC, err := ParseInLoc(s, utc)
	if err != nil {
		t.Fatalf("ParseInLoc(%q, UTC): %v", s, err)
	}

	diff := tShanghai.Unix() - tUTC.Unix()
	wantDiff := int64(-8 * 3600) // Shanghai = UTC+8，同一墙钟时间 Shanghai 比 UTC 早 8 小时
	if diff != wantDiff {
		t.Errorf("Shanghai - UTC = %ds，want %ds", diff, wantDiff)
	}

	// 时区信息正确附着
	if got := tShanghai.Location().String(); got != "Asia/Shanghai" {
		t.Errorf("location = %q, want Asia/Shanghai", got)
	}
}

func TestParseInLoc_DateOnly(t *testing.T) {
	shanghai, _ := time.LoadLocation("Asia/Shanghai")
	t0, err := ParseInLoc("2024-03-15", shanghai)
	if err != nil {
		t.Fatalf("ParseInLoc: %v", err)
	}
	// 纯日期应为当天 00:00:00
	if h, m, s := t0.Clock(); h != 0 || m != 0 || s != 0 {
		t.Errorf("time = %02d:%02d:%02d, want 00:00:00", h, m, s)
	}
	if got := t0.Location().String(); got != "Asia/Shanghai" {
		t.Errorf("location = %q, want Asia/Shanghai", got)
	}
}

func TestParseUnixMilliInLoc(t *testing.T) {
	shanghai, _ := time.LoadLocation("Asia/Shanghai")
	ms, err := ParseUnixMilliInLoc("2024-03-15 10:00:00", shanghai)
	if err != nil {
		t.Fatalf("ParseUnixMilliInLoc: %v", err)
	}
	// 反向验证：还原回 Shanghai 时区应得到原始时间
	got := time.UnixMilli(ms).In(shanghai)
	if got.Hour() != 10 || got.Minute() != 0 || got.Second() != 0 {
		t.Errorf("restored time = %v, want 10:00:00", got)
	}
}

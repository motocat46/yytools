// 版权所有(Copyright)[yangyuan]
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// 作者:  yangyuan
// 创建日期:2026/4/10
package main

import (
	"embed"
	"encoding/json"
	"net/http"
	"sort"
)

//go:embed static
var staticFiles embed.FS

// VisEntry 描述一个可视化页面的完整元数据。
// Pkg 是分组标签（自由字符串，首页按此字段分组）：
//   - yytools 单模块可视化使用 "pkg/<顶级包>" 格式，如 "pkg/algorithms"
//   - 涉及第三方库的对比/评估使用 "benchmarks/<子域>" 格式，如 "benchmarks/ds"
type VisEntry struct {
	Pkg         string `json:"pkg"`
	SubPkg      string `json:"subPkg"`
	Title       string `json:"title"`
	Desc        string `json:"desc"`
	Path        string `json:"path"`
	DataHandler func(http.ResponseWriter, *http.Request) `json:"-"`
}

var registry []VisEntry

// GroupMeta 保存分组的显示元数据，由 RegisterGroup 注册。
type GroupMeta struct {
	Icon string `json:"icon"`
	Desc string `json:"desc"`
}

var groupMeta = map[string]GroupMeta{}

// RegisterGroup 注册分组的图标和描述，由 groups.go 的 init() 调用。
// 同一 pkg 多次调用时后者覆盖前者。
func RegisterGroup(pkg, icon, desc string) {
	groupMeta[pkg] = GroupMeta{Icon: icon, Desc: desc}
}

// Register 注册一个可视化条目，由各 api_*.go 的 init() 调用。
func Register(e VisEntry) {
	registry = append(registry, e)
}

// registryHandler 是统一的 HTTP 入口：
//
//	GET /             → static/index.html
//	GET /chart        → static/chart.html
//	GET /api/registry → 返回所有 VisEntry 的 JSON（不含 DataHandler）
//	GET /api/*        → 查 registry，找到调 DataHandler，否则 404
//	GET /echarts.min.js → static/echarts.min.js
func registryHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		serveStatic(w, r, "static/index.html", "text/html; charset=utf-8")
	case "/chart":
		serveStatic(w, r, "static/chart.html", "text/html; charset=utf-8")
	case "/echarts.min.js":
		serveStatic(w, r, "static/echarts.min.js", "application/javascript")
	case "/api/registry":
		serveRegistry(w)
	default:
		for _, e := range registry {
			if e.Path == r.URL.Path {
				w.Header().Set("Content-Type", "application/json")
				e.DataHandler(w, r)
				return
			}
		}
		http.NotFound(w, r)
	}
}

func serveStatic(w http.ResponseWriter, _ *http.Request, path, contentType string) {
	data, err := staticFiles.ReadFile(path)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", contentType)
	_, _ = w.Write(data)
}

// serveRegistry 按 Pkg 字母排序返回所有条目的 JSON。
func serveRegistry(w http.ResponseWriter) {
	groupMap := map[string][]VisEntry{}
	var groupOrder []string
	for _, e := range registry {
		if _, exists := groupMap[e.Pkg]; !exists {
			groupOrder = append(groupOrder, e.Pkg)
		}
		groupMap[e.Pkg] = append(groupMap[e.Pkg], e)
	}
	sort.Strings(groupOrder)

	type group struct {
		Name    string     `json:"name"`
		Icon    string     `json:"icon"`
		Desc    string     `json:"desc"`
		Entries []VisEntry `json:"entries"`
	}
	groups := make([]group, 0, len(groupOrder))
	for _, name := range groupOrder {
		meta := groupMeta[name]
		groups = append(groups, group{Name: name, Icon: meta.Icon, Desc: meta.Desc, Entries: groupMap[name]})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(groups)
}

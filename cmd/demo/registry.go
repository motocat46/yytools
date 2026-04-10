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
	"html/template"
	"net/http"
	"sort"
)

// VisEntry 描述一个可视化页面的完整元数据。
// Pkg 是分组标签（自由字符串，首页按此字段分组）：
//   - yytools 单模块可视化使用 "pkg/<顶级包>" 格式，如 "pkg/algorithms"
//   - 涉及第三方库的对比/评估使用 "benchmarks/<子域>" 格式，如 "benchmarks/ds"
type VisEntry struct {
	Pkg    string // 分组标签
	SubPkg string // 卡片子包注释，如 "sort/"
	Title  string // 卡片标题
	Desc   string // 一句描述
	Path   string // URL 路径，如 "/sort/efficient"
	Render func(http.ResponseWriter, *http.Request)
}

var registry []VisEntry

// Register 注册一个可视化条目，由各 graph_*.go 的 init() 调用。
func Register(e VisEntry) {
	registry = append(registry, e)
}

// registryHandler 是统一的 HTTP 入口：
// "/" 渲染首页，已注册路径调用对应 Render，其余返回 404。
func registryHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		buildIndexHTML(w)
		return
	}
	for _, e := range registry {
		if e.Path == r.URL.Path {
			e.Render(w, r)
			return
		}
	}
	http.NotFound(w, r)
}

// groupColors 是分组色板，按分组字母排序后顺序取色，超出则循环。
var groupColors = []string{
	"#1a73e8", // 蓝
	"#e67e22", // 橙
	"#27ae60", // 绿
	"#8e44ad", // 紫
	"#c0392b", // 红
	"#16a085", // 青
}

type pkgGroup struct {
	Name    string
	Color   string
	Entries []VisEntry
}

var indexTemplate = template.Must(template.New("index").Parse(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>yytools 可视化</title>
  <style>
    body { font-family: sans-serif; max-width: 960px; margin: 40px auto; padding: 0 20px; color: #333; }
    h1 { font-size: 1.4em; margin-bottom: 1.5em; }
    .group { margin-bottom: 28px; }
    .group-header { display: flex; align-items: center; gap: 10px; margin-bottom: 10px;
                    padding-bottom: 8px; border-bottom: 2px solid var(--c); }
    .group-tag { color: white; border-radius: 4px; padding: 2px 12px; font-size: 11px;
                 font-family: monospace; font-weight: 600; background: var(--c); }
    .cards { display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px; }
    .card { border: 1px solid #e0e0e0; border-radius: 6px; padding: 12px;
            text-decoration: none; display: block; transition: border-color .15s, box-shadow .15s; }
    .card:hover { border-color: var(--c); box-shadow: 0 1px 6px rgba(0,0,0,.1); }
    .card-subpkg { font-size: 10px; color: #999; font-family: monospace; margin-bottom: 4px; }
    .card-title { font-weight: 600; color: var(--c); font-size: 13px; }
    .card-desc { color: #666; font-size: 11px; margin-top: 4px; }
  </style>
</head>
<body>
  <h1>yytools 可视化示例</h1>
  {{range .}}
  <div class="group" style="--c: {{.Color}}">
    <div class="group-header">
      <span class="group-tag">{{.Name}}</span>
    </div>
    <div class="cards">
      {{range .Entries}}
      <a class="card" href="{{.Path}}">
        <div class="card-subpkg">{{.SubPkg}}</div>
        <div class="card-title">{{.Title}}</div>
        <div class="card-desc">{{.Desc}}</div>
      </a>
      {{end}}
    </div>
  </div>
  {{end}}
</body>
</html>`))

// buildIndexHTML 按 Pkg 分组（字母排序）生成首页，写入 w。
func buildIndexHTML(w http.ResponseWriter) {
	groupMap := map[string][]VisEntry{}
	var groupOrder []string
	for _, e := range registry {
		if _, exists := groupMap[e.Pkg]; !exists {
			groupOrder = append(groupOrder, e.Pkg)
		}
		groupMap[e.Pkg] = append(groupMap[e.Pkg], e)
	}
	sort.Strings(groupOrder)

	groups := make([]pkgGroup, 0, len(groupOrder))
	for i, name := range groupOrder {
		groups = append(groups, pkgGroup{
			Name:    name,
			Color:   groupColors[i%len(groupColors)],
			Entries: groupMap[name],
		})
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := indexTemplate.Execute(w, groups); err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
	}
}

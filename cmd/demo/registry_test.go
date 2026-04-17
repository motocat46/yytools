// Package main.
package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// saveRegistry 保存并清空 registry，返回恢复函数。
// 注意：直接操作包级变量，与内部实现耦合；demo 工具可接受，生产代码应提供 Reset() API。
func saveRegistry(t *testing.T) func() {
	t.Helper()
	saved := registry
	registry = nil
	return func() { registry = saved }
}

func TestRegister_AddsEntry(t *testing.T) {
	defer saveRegistry(t)()

	Register(VisEntry{Pkg: "pkg/test", SubPkg: "x/", Title: "T", Desc: "d", Path: "/api/t"})
	if len(registry) != 1 {
		t.Fatalf("got %d entries, want 1", len(registry))
	}
	if registry[0].Path != "/api/t" {
		t.Errorf("got path %q, want /api/t", registry[0].Path)
	}
}

func TestRegistryHandler_KnownPath(t *testing.T) {
	defer saveRegistry(t)()

	Register(VisEntry{
		Pkg: "pkg/test", SubPkg: "x/", Title: "T", Desc: "d",
		Path: "/api/test-page",
		DataHandler: func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{"ok":true}`))
		},
	})

	req := httptest.NewRequest("GET", "/api/test-page", nil)
	rr := httptest.NewRecorder()
	registryHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("got status %d, want 200", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"ok"`) {
		t.Errorf("body missing content, got: %q", rr.Body.String())
	}
}

func TestRegistryHandler_UnknownPath(t *testing.T) {
	defer saveRegistry(t)()

	req := httptest.NewRequest("GET", "/api/no-such-page", nil)
	rr := httptest.NewRecorder()
	registryHandler(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("got status %d, want 404", rr.Code)
	}
}

func TestRegistryHandler_Index(t *testing.T) {
	defer saveRegistry(t)()

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	registryHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "yytools") {
		t.Errorf("index missing yytools in body")
	}
}

// saveGroupMeta 保存并清空 groupMeta，返回恢复函数。
func saveGroupMeta(t *testing.T) func() {
	t.Helper()
	saved := groupMeta
	groupMeta = map[string]GroupMeta{}
	return func() { groupMeta = saved }
}

func TestRegisterGroup_SetsMetadata(t *testing.T) {
	defer saveGroupMeta(t)()

	RegisterGroup("pkg/test", "🔧", "测试描述")

	meta := groupMeta["pkg/test"]
	if meta.Icon != "🔧" {
		t.Errorf("got icon %q, want 🔧", meta.Icon)
	}
	if meta.Desc != "测试描述" {
		t.Errorf("got desc %q, want 测试描述", meta.Desc)
	}
}

func TestRegistryHandler_APIRegistry_IncludesGroupMeta(t *testing.T) {
	defer saveRegistry(t)()
	defer saveGroupMeta(t)()

	Register(VisEntry{Pkg: "pkg/test", SubPkg: "x/", Title: "T", Path: "/api/t",
		DataHandler: func(w http.ResponseWriter, r *http.Request) {}})
	RegisterGroup("pkg/test", "⚡", "测试分组")

	req := httptest.NewRequest("GET", "/api/registry", nil)
	rr := httptest.NewRecorder()
	registryHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rr.Code)
	}
	var groups []struct {
		Name string `json:"name"`
		Icon string `json:"icon"`
		Desc string `json:"desc"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&groups); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("got %d groups, want 1", len(groups))
	}
	if groups[0].Icon != "⚡" {
		t.Errorf("got icon %q, want ⚡", groups[0].Icon)
	}
	if groups[0].Desc != "测试分组" {
		t.Errorf("got desc %q, want 测试分组", groups[0].Desc)
	}
}

func TestRegistryHandler_APIRegistry_GroupsSortedAlphabetically(t *testing.T) {
	defer saveRegistry(t)()

	Register(VisEntry{Pkg: "pkg/ds", SubPkg: "heap/", Title: "堆", Path: "/api/ds/heap"})
	Register(VisEntry{Pkg: "pkg/algorithms", SubPkg: "sort/", Title: "排序", Path: "/api/sort/x"})

	req := httptest.NewRequest("GET", "/api/registry", nil)
	rr := httptest.NewRecorder()
	registryHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rr.Code)
	}

	var groups []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&groups); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(groups) != 2 {
		t.Fatalf("got %d groups, want 2", len(groups))
	}
	if groups[0].Name != "pkg/algorithms" || groups[1].Name != "pkg/ds" {
		t.Errorf("wrong order: %v", groups)
	}
}

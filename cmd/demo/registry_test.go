// Package main.
package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// saveRegistry 保存并清空 registry，返回恢复函数。
func saveRegistry(t *testing.T) func() {
	t.Helper()
	saved := registry
	registry = nil
	return func() { registry = saved }
}

func TestRegister_AddsEntry(t *testing.T) {
	defer saveRegistry(t)()

	Register(VisEntry{Pkg: "pkg/test", SubPkg: "x/", Title: "T", Desc: "d", Path: "/t"})
	if len(registry) != 1 {
		t.Fatalf("got %d entries, want 1", len(registry))
	}
	if registry[0].Path != "/t" {
		t.Errorf("got path %q, want /t", registry[0].Path)
	}
}

func TestRegistryHandler_KnownPath(t *testing.T) {
	defer saveRegistry(t)()

	Register(VisEntry{
		Pkg: "pkg/test", SubPkg: "x/", Title: "T", Desc: "d",
		Path: "/test-page",
		Render: func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("chart-content"))
		},
	})

	req := httptest.NewRequest("GET", "/test-page", nil)
	rr := httptest.NewRecorder()
	registryHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("got status %d, want 200", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "chart-content") {
		t.Errorf("body missing chart-content, got: %q", rr.Body.String())
	}
}

func TestRegistryHandler_UnknownPath(t *testing.T) {
	defer saveRegistry(t)()

	req := httptest.NewRequest("GET", "/no-such-page", nil)
	rr := httptest.NewRecorder()
	registryHandler(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("got status %d, want 404", rr.Code)
	}
}

func TestRegistryHandler_Index_ContainsGroupAndCard(t *testing.T) {
	defer saveRegistry(t)()

	Register(VisEntry{
		Pkg: "pkg/algorithms", SubPkg: "sort/", Title: "排序测试", Desc: "测试描述",
		Path:   "/sort/test",
		Render: func(w http.ResponseWriter, r *http.Request) {},
	})

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	registryHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", rr.Code)
	}
	body := rr.Body.String()
	for _, want := range []string{"pkg/algorithms", "排序测试", "测试描述", "/sort/test"} {
		if !strings.Contains(body, want) {
			t.Errorf("index missing %q in body", want)
		}
	}
}

func TestBuildIndexHTML_GroupsSortedAlphabetically(t *testing.T) {
	defer saveRegistry(t)()

	Register(VisEntry{Pkg: "pkg/ds", SubPkg: "heap/", Title: "堆", Desc: "", Path: "/ds/heap"})
	Register(VisEntry{Pkg: "pkg/algorithms", SubPkg: "sort/", Title: "排序", Desc: "", Path: "/sort/x"})

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	registryHandler(rr, req)

	body := rr.Body.String()
	idxAlgo := strings.Index(body, "pkg/algorithms")
	idxDs := strings.Index(body, "pkg/ds")
	if idxAlgo == -1 || idxDs == -1 {
		t.Fatal("missing group names in index")
	}
	if idxAlgo > idxDs {
		t.Errorf("pkg/algorithms should appear before pkg/ds (alphabetical order)")
	}
}

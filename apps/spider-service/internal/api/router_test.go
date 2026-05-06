// 该文件为路由层的 HTTP 集成测试，使用 httptest 模拟请求验证路由的请求/响应行为，
// 包括创建爬虫、版本管理、registry-auth-refs 和分页列表的正确性。
// 测试使用内存存储（无数据库依赖），通过 NewRouter(service.NewSpiderService()) 构建被测对象。
package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"crawler-platform/apps/spider-service/internal/model"
	"crawler-platform/apps/spider-service/internal/service"
	"github.com/gin-gonic/gin"
)

func TestCreateSpiderReturnsCreatedSpider(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(service.NewSpiderService())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/p1/spiders", strings.NewReader(`{"name":"crawler","language":"go","runtime":"docker","image":"crawler/go:latest","command":["./crawler"]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}

	var spider model.Spider
	if err := json.Unmarshal(w.Body.Bytes(), &spider); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if spider.ID == "" {
		t.Fatal("expected generated id")
	}
	if spider.ProjectID != "p1" || spider.Name != "crawler" || spider.Language != "go" || spider.Runtime != "docker" {
		t.Fatalf("unexpected spider contents: %+v", spider)
	}
	if spider.Image != "crawler/go:latest" {
		t.Fatalf("expected image crawler/go:latest, got %s", spider.Image)
	}
	if len(spider.Command) != 1 || spider.Command[0] != "./crawler" {
		t.Fatalf("expected command to round-trip, got %#v", spider.Command)
	}
	if !strings.Contains(w.Body.String(), `"id":`) || !strings.Contains(w.Body.String(), `"projectId":`) || !strings.Contains(w.Body.String(), `"name":`) || !strings.Contains(w.Body.String(), `"language":`) || !strings.Contains(w.Body.String(), `"runtime":`) || !strings.Contains(w.Body.String(), `"image":`) {
		t.Fatalf("expected lower-case JSON keys, got %s", w.Body.String())
	}
}

func TestCreateSpiderRejectsInvalidInput(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(service.NewSpiderService())
	cases := []struct {
		name string
		body string
	}{
		{name: "invalid language", body: `{"name":"crawler","language":"ruby","runtime":"docker"}`},
		{name: "invalid runtime", body: `{"name":"crawler","language":"go","runtime":"vm"}`},
		{name: "missing image", body: `{"name":"crawler","language":"go","runtime":"docker"}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/p1/spiders", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected status 400, got %d", w.Code)
			}
			if !strings.Contains(w.Body.String(), `"error":`) {
				t.Fatalf("expected error response, got %s", w.Body.String())
			}
		})
	}
}

func TestListSpidersReturnsOnlyRequestedProject(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := service.NewSpiderService()
	spiderP1, err := svc.Create("", "p1", "crawler-a", "go", "docker", "crawler/go-a:latest", []string{"./crawler-a"})
	if err != nil {
		t.Fatalf("expected create success, got error: %v", err)
	}
	spiderP2, err := svc.Create("", "p2", "crawler-b", "python", "host", "", []string{"python", "main.py"})
	if err != nil {
		t.Fatalf("expected create success, got error: %v", err)
	}
	router := NewRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/p1/spiders", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var spiders []model.Spider
	if err := json.Unmarshal(w.Body.Bytes(), &spiders); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(spiders) != 1 {
		t.Fatalf("expected 1 spider, got %d", len(spiders))
	}
	if spiders[0].ID != spiderP1.ID {
		t.Fatalf("expected p1 spider %+v, got %+v", spiderP1, spiders[0])
	}
	if spiders[0].ID == spiderP2.ID {
		t.Fatalf("expected response to exclude p2 spider %+v", spiderP2)
	}
	if !strings.Contains(w.Body.String(), `"projectId":"p1"`) || strings.Contains(w.Body.String(), `"projectId":"p2"`) {
		t.Fatalf("expected response to contain only p1 spider, got %s", w.Body.String())
	}
}

func TestCreateSpiderReturnsFriendlyErrorForBadJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(service.NewSpiderService())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/p1/spiders", strings.NewReader(`not json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "invalid request body") {
		t.Fatalf("expected friendly error message, got %s", w.Body.String())
	}
}

func TestListSpidersWithPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := service.NewSpiderService()
	for i := 0; i < 5; i++ {
		code := "spider-" + string(rune('a'+i))
		if _, err := svc.Create("", "p1", code, "go", "host", "", nil); err != nil {
			t.Fatalf("Create returned error: %v", err)
		}
	}
	router := NewRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/p1/spiders?limit=3&offset=1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var spiders []model.Spider
	if err := json.Unmarshal(w.Body.Bytes(), &spiders); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(spiders) != 3 {
		t.Fatalf("expected 3 spiders with limit=3 offset=1, got %d", len(spiders))
	}
}

func TestCreateVersionAndListVersions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := service.NewSpiderService()
	spider, err := svc.Create("", "p1", "crawler", "go", "docker", "img:latest", nil)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	router := NewRouter(svc)

	// 创建版本
	req := httptest.NewRequest(http.MethodPost, "/api/v1/spiders/"+spider.ID+"/versions", strings.NewReader(`{"version":"v1.0","image":"img:v1","registryAuthRef":"my-registry","command":["./run"]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	// 查询版本列表
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/spiders/"+spider.ID+"/versions", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w2.Code)
	}

	var versions []model.SpiderVersion
	if err := json.Unmarshal(w2.Body.Bytes(), &versions); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(versions) != 1 {
		t.Fatalf("expected 1 version, got %d", len(versions))
	}
	if versions[0].Version != "v1.0" {
		t.Fatalf("expected version v1.0, got %s", versions[0].Version)
	}
}

func TestCreateVersionReturnsNotFoundForMissingSpider(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(service.NewSpiderService())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/spiders/nonexistent/versions", strings.NewReader(`{"version":"v1.0","image":"img:v1"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestListRegistryAuthRefs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(service.NewSpiderService())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/p1/registry-auth-refs", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

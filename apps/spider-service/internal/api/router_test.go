// 该文件为路由层的 HTTP 集成测试，使用 httptest 模拟请求验证路由的请求/响应行为，
// 包括创建爬虫成功/失败场景和按项目 ID 过滤列表的正确性。
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
	spiderP1, err := svc.Create("p1", "crawler-a", "go", "docker", "crawler/go-a:latest", []string{"./crawler-a"})
	if err != nil {
		t.Fatalf("expected create success, got error: %v", err)
	}
	spiderP2, err := svc.Create("p2", "crawler-b", "python", "host", "", []string{"python", "main.py"})
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

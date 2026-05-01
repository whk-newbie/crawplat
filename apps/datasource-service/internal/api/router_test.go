package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"crawler-platform/apps/datasource-service/internal/model"
	"crawler-platform/apps/datasource-service/internal/service"
	"crawler-platform/packages/go-common/httpx"
	"github.com/gin-gonic/gin"
)

func TestCreateDatasourceReturnsCreatedDatasource(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := NewRouter(service.NewDatasourceService())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/datasources", strings.NewReader(`{"projectId":"p1","name":"main","type":"postgresql","config":{"schema":"public"}}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}

	var datasource model.Datasource
	if err := json.Unmarshal(w.Body.Bytes(), &datasource); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if datasource.ID == "" || datasource.Config["schema"] != "public" {
		t.Fatalf("unexpected datasource contents: %#v", datasource)
	}
}

func TestDatasourceLifecycleRoutesUseService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := service.NewDatasourceService()
	created, err := svc.Create("p1", "main", "redis", map[string]string{"db": "0"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	router := NewRouter(svc)

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/datasources?projectId=p1", nil)
	listResp := httptest.NewRecorder()
	router.ServeHTTP(listResp, listReq)
	if listResp.Code != http.StatusOK || !strings.Contains(listResp.Body.String(), `"projectId":"p1"`) {
		t.Fatalf("unexpected list response: %d %s", listResp.Code, listResp.Body.String())
	}
	var listPayload struct {
		Items  []model.Datasource `json:"items"`
		Total  int64              `json:"total"`
		Limit  int                `json:"limit"`
		Offset int                `json:"offset"`
	}
	if err := json.Unmarshal(listResp.Body.Bytes(), &listPayload); err != nil {
		t.Fatalf("failed to decode list response: %v", err)
	}
	if listPayload.Total != 1 || len(listPayload.Items) != 1 {
		t.Fatalf("unexpected list payload: %+v", listPayload)
	}
	if listPayload.Limit != 20 || listPayload.Offset != 0 {
		t.Fatalf("expected default pagination, got limit=%d offset=%d", listPayload.Limit, listPayload.Offset)
	}

	testReq := httptest.NewRequest(http.MethodPost, "/api/v1/datasources/"+created.ID+"/test", nil)
	testResp := httptest.NewRecorder()
	router.ServeHTTP(testResp, testReq)
	if testResp.Code != http.StatusOK || !strings.Contains(testResp.Body.String(), `"status":"ok"`) {
		t.Fatalf("unexpected test response: %d %s", testResp.Code, testResp.Body.String())
	}

	previewReq := httptest.NewRequest(http.MethodPost, "/api/v1/datasources/"+created.ID+"/preview", nil)
	previewResp := httptest.NewRecorder()
	router.ServeHTTP(previewResp, previewReq)
	if previewResp.Code != http.StatusOK || !strings.Contains(previewResp.Body.String(), `"datasourceType":"redis"`) {
		t.Fatalf("unexpected preview response: %d %s", previewResp.Code, previewResp.Body.String())
	}
}

func TestListDatasourcesRespectsPaginationParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := service.NewDatasourceService()
	for i := 0; i < 3; i++ {
		if _, err := svc.Create("p1", "main", "redis", map[string]string{"db": "0"}); err != nil {
			t.Fatalf("Create returned error: %v", err)
		}
	}
	router := NewRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/datasources?projectId=p1&limit=1&offset=1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	var payload httpx.PaginatedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode list response: %v", err)
	}
	if payload.Total != 3 || payload.Limit != 1 || payload.Offset != 1 {
		t.Fatalf("unexpected pagination payload: %+v", payload)
	}
}

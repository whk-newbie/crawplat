package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"crawler-platform/apps/datasource-service/internal/model"
	"crawler-platform/apps/datasource-service/internal/service"
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

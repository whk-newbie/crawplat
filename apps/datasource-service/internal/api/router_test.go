package api

import (
	"context"
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

type stubProber struct {
	testResult    model.TestResult
	previewResult model.PreviewResult
	testErr       error
	previewErr    error
}

func (p *stubProber) Test(_ context.Context, datasource model.Datasource) (model.TestResult, error) {
	if p.testErr != nil {
		return model.TestResult{}, p.testErr
	}
	result := p.testResult
	if result.DatasourceID == "" {
		result.DatasourceID = datasource.ID
	}
	return result, nil
}

func (p *stubProber) Preview(_ context.Context, datasource model.Datasource) (model.PreviewResult, error) {
	if p.previewErr != nil {
		return model.PreviewResult{}, p.previewErr
	}
	result := p.previewResult
	if result.DatasourceID == "" {
		result.DatasourceID = datasource.ID
	}
	if result.DatasourceType == "" {
		result.DatasourceType = datasource.Type
	}
	return result, nil
}

func TestCreateDatasourceReturnsCreatedDatasource(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := service.NewDatasourceService().WithProber(&stubProber{
		testResult:    model.TestResult{Status: "ok", Message: "connection test passed"},
		previewResult: model.PreviewResult{Rows: []map[string]string{{"id": "sample-1"}}},
	})
	router := NewRouter(svc)
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

	svc := service.NewDatasourceService().WithProber(&stubProber{
		testResult:    model.TestResult{Status: "ok", Message: "connection test passed"},
		previewResult: model.PreviewResult{Rows: []map[string]string{{"key": "k1"}}},
	})
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

	svc := service.NewDatasourceService().WithProber(&stubProber{
		testResult:    model.TestResult{Status: "ok", Message: "connection test passed"},
		previewResult: model.PreviewResult{Rows: []map[string]string{{"key": "k1"}}},
	})
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

func TestDatasourceProbeErrorsMapToGatewayResponses(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := service.NewDatasourceService().WithProber(&stubProber{
		testErr:    service.ErrDatasourceConfigInvalid,
		previewErr: service.ErrDatasourceProbeFailed,
	})
	created, err := svc.Create("p1", "main", "redis", map[string]string{"addr": "redis:6379"})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	router := NewRouter(svc)

	testReq := httptest.NewRequest(http.MethodPost, "/api/v1/datasources/"+created.ID+"/test", nil)
	testResp := httptest.NewRecorder()
	router.ServeHTTP(testResp, testReq)
	if testResp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for config error, got %d", testResp.Code)
	}

	previewReq := httptest.NewRequest(http.MethodPost, "/api/v1/datasources/"+created.ID+"/preview", nil)
	previewResp := httptest.NewRecorder()
	router.ServeHTTP(previewResp, previewReq)
	if previewResp.Code != http.StatusBadGateway {
		t.Fatalf("expected 502 for probe error, got %d", previewResp.Code)
	}

	notFoundReq := httptest.NewRequest(http.MethodPost, "/api/v1/datasources/missing/test", nil)
	notFoundResp := httptest.NewRecorder()
	router.ServeHTTP(notFoundResp, notFoundReq)
	if notFoundResp.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing datasource, got %d", notFoundResp.Code)
	}

}

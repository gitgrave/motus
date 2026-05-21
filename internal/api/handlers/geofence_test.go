package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/tamcore/motus/internal/api"
	"github.com/tamcore/motus/internal/api/handlers"
	"github.com/tamcore/motus/internal/model"
	"github.com/tamcore/motus/internal/storage/repository"
	"github.com/tamcore/motus/internal/storage/repository/testutil"
)

const testPolygonGeoJSON = `{"type":"Polygon","coordinates":[[[13.35,52.51],[13.35,52.53],[13.40,52.53],[13.40,52.51],[13.35,52.51]]]}`

func setupGeofenceHandler(t *testing.T) (*handlers.GeofenceHandler, *repository.GeofenceRepository, *model.User) {
	t.Helper()
	pool := testutil.SetupTestDB(t)
	testutil.CleanTables(t, pool)

	userRepo := repository.NewUserRepository(pool)
	geoRepo := repository.NewGeofenceRepository(pool)

	user := &model.User{Email: "geohandler@example.com", PasswordHash: "$2a$10$hash", Name: "Geo Handler"}
	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	h := handlers.NewGeofenceHandler(geoRepo)
	return h, geoRepo, user
}

func TestGeofenceHandler_List_Empty(t *testing.T) {
	h, _, user := setupGeofenceHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/geofences", nil)
	req = withUser(req, user)
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var geofences []model.Geofence
	_ = json.NewDecoder(rr.Body).Decode(&geofences)
	if len(geofences) != 0 {
		t.Errorf("expected empty list, got %d", len(geofences))
	}
}

func TestGeofenceHandler_Create_Success(t *testing.T) {
	h, _, user := setupGeofenceHandler(t)

	body := fmt.Sprintf(`{"name":"Test Fence","description":"A test","geometry":%q}`, testPolygonGeoJSON)
	req := httptest.NewRequest(http.MethodPost, "/api/geofences", bytes.NewReader([]byte(body)))
	req = withUser(req, user)
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d; body: %s", rr.Code, rr.Body.String())
	}

	var g model.Geofence
	_ = json.NewDecoder(rr.Body).Decode(&g)
	if g.Name != "Test Fence" {
		t.Errorf("expected name 'Test Fence', got %q", g.Name)
	}
	if g.ID == 0 {
		t.Error("expected geofence ID to be set")
	}
}

func TestGeofenceHandler_Create_InvalidNameOrDescription(t *testing.T) {
	h, _, user := setupGeofenceHandler(t)

	scriptName := `<script>alert(1)</script>`
	scriptDesc := `<img src=x onerror=alert(1)>`
	longName := strings.Repeat("a", 201)
	longDesc := strings.Repeat("b", 2001)

	tests := []struct {
		name string
		body string
	}{
		{"script in name", fmt.Sprintf(`{"name":%q,"geometry":%q}`, scriptName, testPolygonGeoJSON)},
		{"script in description", fmt.Sprintf(`{"name":"ok","description":%q,"geometry":%q}`, scriptDesc, testPolygonGeoJSON)},
		{"name too long", fmt.Sprintf(`{"name":%q,"geometry":%q}`, longName, testPolygonGeoJSON)},
		{"description too long", fmt.Sprintf(`{"name":"ok","description":%q,"geometry":%q}`, longDesc, testPolygonGeoJSON)},
		{"NUL in name", fmt.Sprintf("{\"name\":\"foo\\u0000bar\",\"geometry\":%q}", testPolygonGeoJSON)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/geofences", bytes.NewReader([]byte(tt.body)))
			req = withUser(req, user)
			rr := httptest.NewRecorder()
			h.Create(rr, req)
			if rr.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d; body: %s", rr.Code, rr.Body.String())
			}
		})
	}
}

func TestGeofenceHandler_Update_InvalidNameOrDescription(t *testing.T) {
	h, geoRepo, user := setupGeofenceHandler(t)

	gf := &model.Geofence{Name: "Valid Fence", Geometry: testPolygonGeoJSON}
	if err := geoRepo.Create(context.Background(), gf); err != nil {
		t.Fatalf("create geofence: %v", err)
	}
	if err := geoRepo.AssociateUser(context.Background(), user.ID, gf.ID); err != nil {
		t.Fatalf("associate user: %v", err)
	}

	tests := []struct {
		name string
		body string
	}{
		{"script in name", `{"name":"<script>x</script>"}`},
		{"script in description", `{"name":"ok","description":"<img src=x>"}`},
		{"name too long", fmt.Sprintf(`{"name":%q}`, strings.Repeat("x", 201))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/geofences/%d", gf.ID), bytes.NewReader([]byte(tt.body)))
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", fmt.Sprintf("%d", gf.ID))
			req = req.WithContext(context.WithValue(api.ContextWithUser(req.Context(), user), chi.RouteCtxKey, rctx))
			rr := httptest.NewRecorder()
			h.Update(rr, req)
			if rr.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d; body: %s", rr.Code, rr.Body.String())
			}
		})
	}
}

func TestGeofenceHandler_Create_MissingName(t *testing.T) {
	h, _, user := setupGeofenceHandler(t)

	body := fmt.Sprintf(`{"geometry":%q}`, testPolygonGeoJSON)
	req := httptest.NewRequest(http.MethodPost, "/api/geofences", bytes.NewReader([]byte(body)))
	req = withUser(req, user)
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestGeofenceHandler_Create_MissingGeometry(t *testing.T) {
	h, _, user := setupGeofenceHandler(t)

	body := `{"name":"No Geometry"}`
	req := httptest.NewRequest(http.MethodPost, "/api/geofences", bytes.NewReader([]byte(body)))
	req = withUser(req, user)
	rr := httptest.NewRecorder()

	h.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestGeofenceHandler_Get_Success(t *testing.T) {
	h, geoRepo, user := setupGeofenceHandler(t)
	ctx := context.Background()

	g := &model.Geofence{Name: "Get Fence", Geometry: testPolygonGeoJSON}
	_ = geoRepo.Create(ctx, g)
	_ = geoRepo.AssociateUser(ctx, user.ID, g.ID)

	req := httptest.NewRequest(http.MethodGet, "/api/geofences/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", fmt.Sprintf("%d", g.ID))
	req = req.WithContext(context.WithValue(api.ContextWithUser(req.Context(), user), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.Get(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
}

func TestGeofenceHandler_Get_Forbidden(t *testing.T) {
	h, geoRepo, user := setupGeofenceHandler(t)
	ctx := context.Background()

	g := &model.Geofence{Name: "No Access Fence", Geometry: testPolygonGeoJSON}
	_ = geoRepo.Create(ctx, g)
	// Do NOT associate user with this geofence.

	req := httptest.NewRequest(http.MethodGet, "/api/geofences/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", fmt.Sprintf("%d", g.ID))
	req = req.WithContext(context.WithValue(api.ContextWithUser(req.Context(), user), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.Get(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestGeofenceHandler_Update_Success(t *testing.T) {
	h, geoRepo, user := setupGeofenceHandler(t)
	ctx := context.Background()

	g := &model.Geofence{Name: "Before", Geometry: testPolygonGeoJSON}
	_ = geoRepo.Create(ctx, g)
	_ = geoRepo.AssociateUser(ctx, user.ID, g.ID)

	body := `{"name":"After"}`
	req := httptest.NewRequest(http.MethodPut, "/api/geofences/1", bytes.NewReader([]byte(body)))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", fmt.Sprintf("%d", g.ID))
	req = req.WithContext(context.WithValue(api.ContextWithUser(req.Context(), user), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.Update(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; body: %s", rr.Code, rr.Body.String())
	}
}

func TestGeofenceHandler_Delete_Success(t *testing.T) {
	h, geoRepo, user := setupGeofenceHandler(t)
	ctx := context.Background()

	g := &model.Geofence{Name: "Delete Me", Geometry: testPolygonGeoJSON}
	_ = geoRepo.Create(ctx, g)
	_ = geoRepo.AssociateUser(ctx, user.ID, g.ID)

	req := httptest.NewRequest(http.MethodDelete, "/api/geofences/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", fmt.Sprintf("%d", g.ID))
	req = req.WithContext(context.WithValue(api.ContextWithUser(req.Context(), user), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.Delete(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
}

func TestGeofenceHandler_InvalidID(t *testing.T) {
	h, _, user := setupGeofenceHandler(t)

	endpoints := []struct {
		name   string
		method string
		path   string
		body   string
		fn     func(http.ResponseWriter, *http.Request)
	}{
		{"Get", http.MethodGet, "/api/geofences/abc", "", h.Get},
		{"Update", http.MethodPut, "/api/geofences/abc", `{"name":"x"}`, h.Update},
		{"Delete", http.MethodDelete, "/api/geofences/abc", "", h.Delete},
	}

	for _, tt := range endpoints {
		t.Run(tt.name, func(t *testing.T) {
			var bodyReader *bytes.Reader
			if tt.body != "" {
				bodyReader = bytes.NewReader([]byte(tt.body))
			}
			var req *http.Request
			if bodyReader != nil {
				req = httptest.NewRequest(tt.method, tt.path, bodyReader)
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", "abc")
			req = req.WithContext(context.WithValue(api.ContextWithUser(req.Context(), user), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			tt.fn(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d", rr.Code)
			}
		})
	}
}

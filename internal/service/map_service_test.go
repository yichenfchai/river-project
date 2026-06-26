package service

import (
	"context"
	"testing"

	"github.com/yichenfchai/river-project/internal/model"
	apperrors "github.com/yichenfchai/river-project/pkg/errors"
)

// ─── Mock Map Repository ───

type mockMapRepo struct {
	listLayersFn   func(ctx context.Context) ([]model.MapLayer, error)
	getLayerFn     func(ctx context.Context, id string) (*model.MapLayer, error)
	listPOIsFn     func(ctx context.Context, lat, lng, radius float64, category string) ([]model.MapPOI, error)
	getPOIFn       func(ctx context.Context, id string) (*model.MapPOI, error)
	searchPOIsFn   func(ctx context.Context, keyword string) ([]model.MapPOI, error)
}

func (m *mockMapRepo) ListLayers(ctx context.Context) ([]model.MapLayer, error) {
	if m.listLayersFn != nil {
		return m.listLayersFn(ctx)
	}
	return nil, nil
}
func (m *mockMapRepo) GetLayer(ctx context.Context, id string) (*model.MapLayer, error) {
	if m.getLayerFn != nil {
		return m.getLayerFn(ctx, id)
	}
	return nil, apperrors.NewDefault(apperrors.ErrNotFound)
}
func (m *mockMapRepo) ListPOIs(ctx context.Context, lat, lng, radius float64, category string) ([]model.MapPOI, error) {
	if m.listPOIsFn != nil {
		return m.listPOIsFn(ctx, lat, lng, radius, category)
	}
	return nil, nil
}
func (m *mockMapRepo) Seed(ctx context.Context) error { return nil }
func (m *mockMapRepo) GetPOI(ctx context.Context, id string) (*model.MapPOI, error) {
	if m.getPOIFn != nil {
		return m.getPOIFn(ctx, id)
	}
	return nil, apperrors.NewDefault(apperrors.ErrPOINotFound)
}
func (m *mockMapRepo) SearchPOIs(ctx context.Context, keyword string) ([]model.MapPOI, error) {
	if m.searchPOIsFn != nil {
		return m.searchPOIsFn(ctx, keyword)
	}
	return nil, nil
}

// ─── Tests ───

func TestMapService_ListLayers(t *testing.T) {
	ctx := context.Background()
	repo := &mockMapRepo{
		listLayersFn: func(ctx context.Context) ([]model.MapLayer, error) {
			return []model.MapLayer{
				{ID: "l1", Name: "元代运河", Era: "元", Color: "#c0392b", SortOrder: 1},
				{ID: "l2", Name: "明代运河", Era: "明", Color: "#2980b9", SortOrder: 2},
				{ID: "l3", Name: "清代运河", Era: "清", Color: "#27ae60", SortOrder: 3},
				{ID: "l4", Name: "现代运河", Era: "现代", Color: "#8e44ad", SortOrder: 4},
			}, nil
		},
	}
	svc := NewMapService(repo)

	layers, err := svc.ListLayers(ctx)
	if err != nil {
		t.Fatalf("ListLayers error: %v", err)
	}
	if len(layers) != 4 {
		t.Fatalf("len = %d, want 4", len(layers))
	}

	tests := []struct {
		idx  int
		name string
		era  string
	}{
		{0, "元代运河", "元"},
		{1, "明代运河", "明"},
		{2, "清代运河", "清"},
		{3, "现代运河", "现代"},
	}
	for _, tt := range tests {
		if layers[tt.idx].Name != tt.name {
			t.Errorf("Layer[%d].Name = %q, want %q", tt.idx, layers[tt.idx].Name, tt.name)
		}
		if layers[tt.idx].Era != tt.era {
			t.Errorf("Layer[%d].Era = %q, want %q", tt.idx, layers[tt.idx].Era, tt.era)
		}
	}
}

func TestMapService_ListLayers_Empty(t *testing.T) {
	ctx := context.Background()
	repo := &mockMapRepo{
		listLayersFn: func(ctx context.Context) ([]model.MapLayer, error) {
			return []model.MapLayer{}, nil
		},
	}
	svc := NewMapService(repo)

	layers, err := svc.ListLayers(ctx)
	if err != nil {
		t.Fatalf("ListLayers error: %v", err)
	}
	if len(layers) != 0 {
		t.Errorf("Expected empty result, got %d layers", len(layers))
	}
}

func TestMapService_GetLayer(t *testing.T) {
	ctx := context.Background()
	repo := &mockMapRepo{
		getLayerFn: func(ctx context.Context, id string) (*model.MapLayer, error) {
			return &model.MapLayer{
				ID: "l1", Name: "元代运河", Era: "元",
				Description: "元世祖忽必烈时期开凿的京杭大运河",
				Color:       "#c0392b",
				GeoJSON:     `{"type":"FeatureCollection","features":[]}`,
			}, nil
		},
	}
	svc := NewMapService(repo)

	layer, err := svc.GetLayer(ctx, "l1")
	if err != nil {
		t.Fatalf("GetLayer error: %v", err)
	}
	if layer.Name != "元代运河" {
		t.Errorf("Name = %q", layer.Name)
	}
	if layer.GeoJSON == nil {
		t.Error("GeoJSON should not be nil")
	}
}

func TestMapService_GetLayer_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := &mockMapRepo{} // default returns ErrNotFound
	svc := NewMapService(repo)

	_, err := svc.GetLayer(ctx, "nonexistent")
	if err == nil {
		t.Fatal("Expected error for nonexistent layer")
	}
}

func TestMapService_ListPOIs(t *testing.T) {
	ctx := context.Background()
	repo := &mockMapRepo{
		listPOIsFn: func(ctx context.Context, lat, lng, radius float64, category string) ([]model.MapPOI, error) {
			return []model.MapPOI{
				{ID: "p1", Name: "扬州运河大桥", Category: "桥梁", Lat: 32.39, Lng: 119.42, LayerID: "l1"},
				{ID: "p2", Name: "苏州运河码头", Category: "港口", Lat: 31.30, Lng: 120.62, LayerID: "l1"},
			}, nil
		},
	}
	svc := NewMapService(repo)

	pois, err := svc.ListPOIs(ctx, 32.0, 119.0, 50.0, "")
	if err != nil {
		t.Fatalf("ListPOIs error: %v", err)
	}
	if len(pois) != 2 {
		t.Fatalf("len = %d, want 2", len(pois))
	}
	if pois[0].Category != "桥梁" {
		t.Errorf("POI[0].Category = %q", pois[0].Category)
	}
}

func TestMapService_ListPOIs_FilterByCategory(t *testing.T) {
	ctx := context.Background()
	var capturedCategory string
	repo := &mockMapRepo{
		listPOIsFn: func(ctx context.Context, lat, lng, radius float64, category string) ([]model.MapPOI, error) {
			capturedCategory = category
			return []model.MapPOI{
				{ID: "p1", Name: "古桥", Category: "桥梁"},
			}, nil
		},
	}
	svc := NewMapService(repo)

	_, err := svc.ListPOIs(ctx, 30.0, 120.0, 10.0, "桥梁")
	if err != nil {
		t.Fatalf("ListPOIs error: %v", err)
	}
	if capturedCategory != "桥梁" {
		t.Errorf("Category = %q, want %q (filter not passed through)", capturedCategory, "桥梁")
	}
}

func TestMapService_ListPOIs_Empty(t *testing.T) {
	ctx := context.Background()
	repo := &mockMapRepo{
		listPOIsFn: func(ctx context.Context, lat, lng, radius float64, category string) ([]model.MapPOI, error) {
			return []model.MapPOI{}, nil
		},
	}
	svc := NewMapService(repo)

	pois, err := svc.ListPOIs(ctx, 10.0, 10.0, 5.0, "寺庙")
	if err != nil {
		t.Fatalf("ListPOIs error: %v", err)
	}
	if len(pois) != 0 {
		t.Errorf("Expected empty result, got %d", len(pois))
	}
}

package service

import (
	"context"

	"github.com/yichenfchai/river-project/internal/repository"
)

// MapService 地图服务接口
type MapService interface {
	ListLayers(ctx context.Context) ([]LayerInfo, error)
	GetLayer(ctx context.Context, id string) (*LayerDetail, error)
	ListPOIs(ctx context.Context, lat, lng, radius float64, category string) ([]POIInfo, error)
}

type LayerInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Era         string `json:"era"`
	Description string `json:"description"`
	Color       string `json:"color"`
	SortOrder   int    `json:"sort_order"`
}

type LayerDetail struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Era         string `json:"era"`
	Description string `json:"description"`
	Color       string `json:"color"`
	GeoJSON     any    `json:"geojson"`
}

type POIInfo struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	LayerID     string  `json:"layer_id"`
}

type mapService struct {
	repo repository.MapRepository
}

func NewMapService(repo repository.MapRepository) MapService {
	return &mapService{repo: repo}
}

func (s *mapService) ListLayers(ctx context.Context) ([]LayerInfo, error) {
	layers, err := s.repo.ListLayers(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]LayerInfo, len(layers))
	for i, l := range layers {
		result[i] = LayerInfo{
			ID: l.ID, Name: l.Name, Era: l.Era,
			Description: l.Description, Color: l.Color, SortOrder: l.SortOrder,
		}
	}
	return result, nil
}

func (s *mapService) GetLayer(ctx context.Context, id string) (*LayerDetail, error) {
	layer, err := s.repo.GetLayer(ctx, id)
	if err != nil {
		return nil, err
	}
	return &LayerDetail{
		ID: layer.ID, Name: layer.Name, Era: layer.Era,
		Description: layer.Description, Color: layer.Color,
		GeoJSON: layer.GeoJSON,
	}, nil
}

func (s *mapService) ListPOIs(ctx context.Context, lat, lng, radius float64, category string) ([]POIInfo, error) {
	pois, err := s.repo.ListPOIs(ctx, lat, lng, radius, category)
	if err != nil {
		return nil, err
	}
	result := make([]POIInfo, len(pois))
	for i, p := range pois {
		result[i] = POIInfo{
			ID: p.ID, Name: p.Name, Description: p.Description,
			Category: p.Category, Lat: p.Lat, Lng: p.Lng, LayerID: p.LayerID,
		}
	}
	return result, nil
}

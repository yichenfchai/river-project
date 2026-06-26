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
	GetPOI(ctx context.Context, id string) (*POIInfo, error)
	SearchPOIs(ctx context.Context, keyword string) ([]POIInfo, error)
	GetTimeline(ctx context.Context, yearFrom, yearTo int) ([]TimelineEvent, error)
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
	IconURL     string  `json:"icon_url"`
}

type TimelineEvent struct {
	Year        int    `json:"year"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Era         string `json:"era"`
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
			Category: p.Category, Lat: p.Lat, Lng: p.Lng,
			LayerID: p.LayerID, IconURL: p.IconURL,
		}
	}
	return result, nil
}

func (s *mapService) GetPOI(ctx context.Context, id string) (*POIInfo, error) {
	poi, err := s.repo.GetPOI(ctx, id)
	if err != nil {
		return nil, err
	}
	return &POIInfo{
		ID: poi.ID, Name: poi.Name, Description: poi.Description,
		Category: poi.Category, Lat: poi.Lat, Lng: poi.Lng,
		LayerID: poi.LayerID, IconURL: poi.IconURL,
	}, nil
}

func (s *mapService) SearchPOIs(ctx context.Context, keyword string) ([]POIInfo, error) {
	pois, err := s.repo.SearchPOIs(ctx, keyword)
	if err != nil {
		return nil, err
	}
	result := make([]POIInfo, len(pois))
	for i, p := range pois {
		result[i] = POIInfo{
			ID: p.ID, Name: p.Name, Description: p.Description,
			Category: p.Category, Lat: p.Lat, Lng: p.Lng,
			LayerID: p.LayerID, IconURL: p.IconURL,
		}
	}
	if result == nil {
		result = []POIInfo{}
	}
	return result, nil
}

func (s *mapService) GetTimeline(ctx context.Context, yearFrom, yearTo int) ([]TimelineEvent, error) {
	events := []TimelineEvent{
		{Year: -486, Title: "吴王夫差开凿邗沟", Description: "春秋时期吴王夫差为北上争霸开凿邗沟，贯通长江与淮河，被认为是大运河最早的一段", Era: "春秋"},
		{Year: 605, Title: "隋炀帝开凿通济渠", Description: "隋炀帝征发百万民夫开凿通济渠，连接黄河与淮河，奠定隋唐大运河基础", Era: "隋"},
		{Year: 608, Title: "开凿永济渠", Description: "隋朝开凿永济渠，北达涿郡（今北京），使大运河北段延伸", Era: "隋"},
		{Year: 610, Title: "开凿江南运河", Description: "隋朝整治江南运河，从镇江经苏州到杭州，完成南北大运河贯通", Era: "隋"},
		{Year: 1271, Title: "元朝裁弯取直", Description: "元朝定都大都（北京），裁弯取直大运河，不再绕道洛阳，京杭大运河正式形成", Era: "元"},
		{Year: 1293, Title: "通惠河完工", Description: "郭守敬主持修建通惠河，大运河直接从杭州通至北京城内的积水潭", Era: "元"},
		{Year: 1411, Title: "明代重修运河", Description: "明成祖迁都北京后，全面整治京杭大运河漕运体系", Era: "明"},
		{Year: 1855, Title: "黄河改道", Description: "黄河在铜瓦厢决口改道，冲断运河山东段，运河漕运开始衰落", Era: "清"},
		{Year: 1901, Title: "漕运废止", Description: "清政府宣布停止漕运，千年运河航运史告一段落", Era: "清"},
		{Year: 2002, Title: "南水北调东线开工", Description: "南水北调东线工程全面开工，利用京杭大运河作为输水干线", Era: "现代"},
		{Year: 2014, Title: "入选世界遗产", Description: "中国大运河项目成功入选《世界遗产名录》，成为我国第46个世界遗产", Era: "现代"},
		{Year: 2020, Title: "大运河文化带建设", Description: "大运河国家文化公园建设全面启动，打造中华文化重要标志", Era: "现代"},
	}

	var filtered []TimelineEvent
	for _, e := range events {
		if (yearFrom == 0 || e.Year >= yearFrom) && (yearTo == 0 || e.Year <= yearTo) {
			filtered = append(filtered, e)
		}
	}
	if filtered == nil {
		filtered = []TimelineEvent{}
	}
	return filtered, nil
}

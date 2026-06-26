package repository

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/yichenfchai/river-project/internal/model"
	apperrors "github.com/yichenfchai/river-project/pkg/errors"
)

// MapRepository 地图数据访问接口
type MapRepository interface {
	ListLayers(ctx context.Context) ([]model.MapLayer, error)
	GetLayer(ctx context.Context, id string) (*model.MapLayer, error)
	ListPOIs(ctx context.Context, lat, lng, radius float64, category string) ([]model.MapPOI, error)
	GetPOI(ctx context.Context, id string) (*model.MapPOI, error)
	SearchPOIs(ctx context.Context, keyword string) ([]model.MapPOI, error)
	Seed(ctx context.Context) error
}

type mapRepo struct {
	db *gorm.DB
}

func NewMapRepo(db *gorm.DB) MapRepository {
	return &mapRepo{db: db}
}

func (r *mapRepo) ListLayers(ctx context.Context) ([]model.MapLayer, error) {
	var layers []model.MapLayer
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("sort_order ASC").
		Find(&layers).Error
	if err != nil {
		return nil, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return layers, nil
}

func (r *mapRepo) GetLayer(ctx context.Context, id string) (*model.MapLayer, error) {
	var layer model.MapLayer
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&layer).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewDefault(apperrors.ErrPOINotFound)
		}
		return nil, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return &layer, nil
}

// ListPOIs 按中心点+半径查询 POI（简化版，可用 bounding box 代替 PostGIS）
func (r *mapRepo) ListPOIs(ctx context.Context, lat, lng, radius float64, category string) ([]model.MapPOI, error) {
	query := r.db.WithContext(ctx).Model(&model.MapPOI{})

	// 简易 bounding box：纬度每度约 111km，经度每度约 111km * cos(lat)
	latDelta := radius / 111000.0
	lngDelta := radius / (111000.0 * cos(lat))

	query = query.Where("lat BETWEEN ? AND ?", lat-latDelta, lat+latDelta).
		Where("lng BETWEEN ? AND ?", lng-lngDelta, lng+lngDelta)

	if category != "" {
		query = query.Where("category = ?", category)
	}

	var pois []model.MapPOI
	if err := query.Find(&pois).Error; err != nil {
		return nil, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return pois, nil
}

func cos(deg float64) float64 {
	const degToRad = 0.017453292519943295
	rad := deg * degToRad
	// Taylor approximation of cos(x)
	x := 1.0
	term := 1.0
	for i := 1; i <= 6; i++ {
		term = -term * rad * rad / float64((2*i-1)*(2*i))
		x += term
	}
	if x < 0.01 {
		x = 0.01
	}
	return x
}

func (r *mapRepo) GetPOI(ctx context.Context, id string) (*model.MapPOI, error) {
	var poi model.MapPOI
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&poi).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewDefault(apperrors.ErrPOINotFound)
		}
		return nil, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return &poi, nil
}

func (r *mapRepo) SearchPOIs(ctx context.Context, keyword string) ([]model.MapPOI, error) {
	var pois []model.MapPOI
	like := "%" + keyword + "%"
	err := r.db.WithContext(ctx).
		Where("name LIKE ? OR description LIKE ? OR category LIKE ?", like, like, like).
		Order("name ASC").Limit(20).
		Find(&pois).Error
	if err != nil {
		return nil, apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
	}
	return pois, nil
}

// Seed 幂等种子数据：4 个运河图层 + 15 个文化遗产 POI
func (r *mapRepo) Seed(ctx context.Context) error {
	seedLayers := []model.MapLayer{
		{
			ID: "map-seed-sui-tang", Name: "隋唐运河", Era: "sui-tang",
			Description: "隋唐大运河以洛阳为中心，北至涿郡（今北京），南达余杭（今杭州）",
			Color: "#e6a23c", SortOrder: 1, IsActive: true,
			GeoJSON: `{"type":"LineString","coordinates":[[116.4,39.9],[116.2,39.0],[115.9,38.0],[116.3,37.4],[116.6,35.4],[117.2,34.3],[119.1,33.5],[112.4,34.6],[114.3,32.1],[116.0,31.0],[119.4,32.4],[120.6,31.3],[120.2,30.3]]}`,
		},
		{
			ID: "map-seed-yuan-ming-qing", Name: "元明清运河", Era: "yuan-ming-qing",
			Description: "京杭大运河直通南北，从北京直达杭州，全长约1794公里",
			Color: "#409eff", SortOrder: 2, IsActive: true,
			GeoJSON: `{"type":"LineString","coordinates":[[116.4,39.9],[117.2,39.1],[116.8,38.3],[116.3,37.4],[117.0,36.7],[116.6,35.4],[117.2,34.3],[118.1,33.9],[119.1,33.5],[119.4,32.4],[119.4,32.2],[119.6,31.8],[120.6,31.3],[120.2,30.3]]}`,
		},
		{
			ID: "map-seed-south-north", Name: "现代南水北调东线", Era: "modern",
			Description: "南水北调东线工程，沿京杭大运河线路从扬州向华北调水",
			Color: "#67c23a", SortOrder: 3, IsActive: true,
			GeoJSON: `{"type":"LineString","coordinates":[[119.4,32.4],[119.4,32.2],[119.0,33.5],[117.2,34.3],[116.6,35.4],[116.3,37.4],[116.8,38.3],[117.2,39.1]]}`,
		},
		{
			ID: "map-seed-eco-monitor", Name: "生态监测站点", Era: "modern",
			Description: "大运河沿线水质和生态监测站分布",
			Color: "#f56c6c", SortOrder: 4, IsActive: true,
			GeoJSON: `{"type":"FeatureCollection","features":[]}`,
		},
	}

	for _, layer := range seedLayers {
		if err := r.db.WithContext(ctx).
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(&layer).Error; err != nil {
			return apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
		}
	}

	seedPOIs := []model.MapPOI{
		{ID: "poi-seed-001", LayerID: "map-seed-yuan-ming-qing", Name: "扬州古运河渡口", Description: "唐代诗人李白曾在此登船，留下千古名句", Category: "cultural", Lat: 32.39, Lng: 119.43, IconURL: ""},
		{ID: "poi-seed-002", LayerID: "map-seed-yuan-ming-qing", Name: "苏州山塘河", Description: "白居易任苏州刺史时开凿，七里山塘繁华至今", Category: "cultural", Lat: 31.32, Lng: 120.62, IconURL: ""},
		{ID: "poi-seed-003", LayerID: "map-seed-yuan-ming-qing", Name: "北京通州运河码头", Description: "元明清三代漕运北端终点，万国来朝第一站", Category: "cultural", Lat: 39.90, Lng: 116.66, IconURL: ""},
		{ID: "poi-seed-004", LayerID: "map-seed-yuan-ming-qing", Name: "山东南旺分水枢纽", Description: "运河之心，明代水利奇迹，解决运河最高点水源问题", Category: "cultural", Lat: 35.56, Lng: 116.35, IconURL: ""},
		{ID: "poi-seed-005", LayerID: "map-seed-yuan-ming-qing", Name: "镇江谏壁船闸", Description: "长江与运河交汇处，古代江南运河的北端门户", Category: "cultural", Lat: 32.17, Lng: 119.45, IconURL: ""},
		{ID: "poi-seed-006", LayerID: "map-seed-yuan-ming-qing", Name: "杭州拱宸桥", Description: "京杭大运河南端终点标志，杭州现存最高最长的古桥", Category: "cultural", Lat: 30.32, Lng: 120.14, IconURL: ""},
		{ID: "poi-seed-007", LayerID: "map-seed-yuan-ming-qing", Name: "天津三岔河口", Description: "南运河、北运河与海河交汇处，天津城市发源地", Category: "cultural", Lat: 39.14, Lng: 117.19, IconURL: ""},
		{ID: "poi-seed-008", LayerID: "map-seed-yuan-ming-qing", Name: "临清钞关", Description: "明代运河沿线七大钞关之首，见证了运河商贸繁荣", Category: "cultural", Lat: 36.85, Lng: 115.71, IconURL: ""},
		{ID: "poi-seed-009", LayerID: "map-seed-sui-tang", Name: "洛阳含嘉仓遗址", Description: "隋唐时期的大型皇家粮仓，可储粮五百万石", Category: "cultural", Lat: 34.68, Lng: 112.44, IconURL: ""},
		{ID: "poi-seed-010", LayerID: "map-seed-sui-tang", Name: "汴州（开封）运河枢纽", Description: "隋唐大运河通济渠段的核心节点", Category: "cultural", Lat: 34.79, Lng: 114.31, IconURL: ""},
		{ID: "poi-seed-011", LayerID: "map-seed-eco-monitor", Name: "淮安水质监测站", Description: "国家地表水自动监测站，实时监测运河水质六大指标", Category: "ecology", Lat: 33.51, Lng: 119.14, IconURL: ""},
		{ID: "poi-seed-012", LayerID: "map-seed-eco-monitor", Name: "济宁生态观测站", Description: "南四湖流域生态修复示范区，鸟类栖息地监测", Category: "ecology", Lat: 35.38, Lng: 116.58, IconURL: ""},
		{ID: "poi-seed-013", LayerID: "map-seed-eco-monitor", Name: "德州湿地保护区", Description: "运河德州段的芦苇湿地生态系统保护区", Category: "ecology", Lat: 37.43, Lng: 116.36, IconURL: ""},
		{ID: "poi-seed-014", LayerID: "map-seed-eco-monitor", Name: "扬州三湾湿地公园", Description: "运河水生态修复示范工程，引江水利枢纽所在地", Category: "ecology", Lat: 32.38, Lng: 119.42, IconURL: ""},
		{ID: "poi-seed-015", LayerID: "map-seed-eco-monitor", Name: "沧州南运河生态廊道", Description: "大运河文化带生态修复重点项目，全长约54公里", Category: "ecology", Lat: 38.30, Lng: 116.84, IconURL: ""},
	}

	for _, poi := range seedPOIs {
		if err := r.db.WithContext(ctx).
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(&poi).Error; err != nil {
			return apperrors.WrapDefault(apperrors.ErrDatabaseError, err)
		}
	}

	return nil
}

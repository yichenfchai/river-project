package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/your-org/grand-canal-guardian/internal/service"
	"github.com/your-org/grand-canal-guardian/pkg/errors"
	"github.com/your-org/grand-canal-guardian/pkg/response"
)

type MapHandler struct {
	svc service.MapService
}

func NewMapHandler(svc service.MapService) *MapHandler {
	return &MapHandler{svc: svc}
}

// ListLayers 获取可用图层列表
func (h *MapHandler) ListLayers(c *gin.Context) {
	layers, err := h.svc.ListLayers(c.Request.Context())
	if err != nil {
		response.Error(c, toAppError(err))
		return
	}
	response.OK(c, layers)
}

// GetLayer 获取单个图层（含 GeoJSON）
func (h *MapHandler) GetLayer(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, errors.BadRequest("缺少图层 ID"))
		return
	}

	layer, err := h.svc.GetLayer(c.Request.Context(), id)
	if err != nil {
		response.Error(c, toAppError(err))
		return
	}
	response.OK(c, layer)
}

// ListPOIs 查询周边 POI
func (h *MapHandler) ListPOIs(c *gin.Context) {
	lat, err := strconv.ParseFloat(c.Query("lat"), 64)
	if err != nil || lat < -90 || lat > 90 {
		response.Error(c, errors.BadRequest("lat 参数无效"))
		return
	}
	lng, err := strconv.ParseFloat(c.Query("lng"), 64)
	if err != nil || lng < -180 || lng > 180 {
		response.Error(c, errors.BadRequest("lng 参数无效"))
		return
	}
	radius, err := strconv.ParseFloat(c.DefaultQuery("radius", "50000"), 64)
	if err != nil || radius <= 0 {
		radius = 50000
	}

	pois, err := h.svc.ListPOIs(c.Request.Context(), lat, lng, radius, c.Query("category"))
	if err != nil {
		response.Error(c, toAppError(err))
		return
	}
	response.OK(c, pois)
}

package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/yichenfchai/river-project/internal/service"
	"github.com/yichenfchai/river-project/pkg/errors"
	"github.com/yichenfchai/river-project/pkg/response"
)

type GeoHandler struct {
	svc service.GeoService
}

func NewGeoHandler(svc service.GeoService) *GeoHandler {
	return &GeoHandler{svc: svc}
}

func (h *GeoHandler) CanalIntro(c *gin.Context) {
	ip := c.ClientIP()

	result, err := h.svc.GetCanalIntro(c.Request.Context(), ip)
	if err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.OK(c, result)
}

func (h *GeoHandler) CanalIntroByCoords(c *gin.Context) {
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

	result, err := h.svc.GetCanalIntroByCoords(c.Request.Context(), lat, lng)
	if err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.OK(c, result)
}

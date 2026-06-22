package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/yichenfchai/river-project/internal/service"
	"github.com/yichenfchai/river-project/pkg/auth"
	"github.com/yichenfchai/river-project/pkg/errors"
	"github.com/yichenfchai/river-project/pkg/response"
)

type ShopHandler struct {
	svc service.ShopService
}

func NewShopHandler(svc service.ShopService) *ShopHandler {
	return &ShopHandler{svc: svc}
}

func (h *ShopHandler) ListItems(c *gin.Context) {
	items, err := h.svc.ListItems(c.Request.Context())
	if err != nil {
		response.Error(c, toAppError(err))
		return
	}
	response.OK(c, items)
}

type redeemRequest struct {
	ItemID string `json:"item_id" binding:"required"`
}

func (h *ShopHandler) Redeem(c *gin.Context) {
	var req redeemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.BadRequest("请选择要兑换的商品"))
		return
	}

	userID := auth.GetUserID(c)
	role := auth.GetRole(c)
	if userID == "" {
		response.Error(c, errors.NewDefault(errors.ErrUnauthorized))
		return
	}

	result, err := h.svc.Redeem(c.Request.Context(), userID, role, req.ItemID)
	if err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.OK(c, result)
}

func (h *ShopHandler) GetHistory(c *gin.Context) {
	userID := auth.GetUserID(c)
	if userID == "" {
		response.Error(c, errors.NewDefault(errors.ErrUnauthorized))
		return
	}

	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "page_size", 20)

	result, err := h.svc.GetHistory(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.OK(c, gin.H{
		"items": result.Items,
		"pagination": gin.H{
			"page":  result.Page,
			"total": result.Total,
		},
	})
}

func queryInt(c *gin.Context, key string, fallback int) int {
	val := c.Query(key)
	if val == "" {
		return fallback
	}
	n := 0
	for _, ch := range val {
		if ch < '0' || ch > '9' {
			return fallback
		}
		n = n*10 + int(ch-'0')
	}
	if n < 1 {
		return fallback
	}
	return n
}

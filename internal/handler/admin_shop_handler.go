package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/yichenfchai/river-project/internal/model"
	"github.com/yichenfchai/river-project/internal/service"
	"github.com/yichenfchai/river-project/pkg/errors"
	"github.com/yichenfchai/river-project/pkg/response"
)

// AdminShopHandler 管理端 — 商品 CRUD
type AdminShopHandler struct {
	svc service.ShopService
}

func NewAdminShopHandler(svc service.ShopService) *AdminShopHandler {
	return &AdminShopHandler{svc: svc}
}

type createItemRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=128"`
	Description string `json:"description" binding:"max=512"`
	ImageURL    string `json:"image_url" binding:"max=512"`
	PointsCost  int    `json:"points_cost" binding:"required,min=1"`
	Stock       int    `json:"stock"`
}

func (h *AdminShopHandler) CreateItem(c *gin.Context) {
	var req createItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.BadRequest("请正确填写商品信息"))
		return
	}

	item := &model.ShopItem{
		Name:        req.Name,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		PointsCost:  req.PointsCost,
		Stock:       req.Stock,
		IsActive:    true,
	}

	if err := h.svc.CreateItem(c.Request.Context(), item); err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.Created(c, item)
}

type updateItemRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=128"`
	Description string `json:"description" binding:"max=512"`
	ImageURL    string `json:"image_url" binding:"max=512"`
	PointsCost  int    `json:"points_cost" binding:"required,min=1"`
	Stock       int    `json:"stock"`
	IsActive    *bool  `json:"is_active" binding:"required"`
}

func (h *AdminShopHandler) UpdateItem(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, errors.BadRequest("缺少商品 ID"))
		return
	}

	var req updateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.BadRequest("请正确填写商品信息"))
		return
	}

	item := &model.ShopItem{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		PointsCost:  req.PointsCost,
		Stock:       req.Stock,
		IsActive:    *req.IsActive,
	}

	if err := h.svc.UpdateItem(c.Request.Context(), item); err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.OK(c, item)
}

func (h *AdminShopHandler) DeleteItem(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.Error(c, errors.BadRequest("缺少商品 ID"))
		return
	}

	if err := h.svc.DeleteItem(c.Request.Context(), id); err != nil {
		response.Error(c, toAppError(err))
		return
	}

	response.OKMessage(c, "已删除")
}

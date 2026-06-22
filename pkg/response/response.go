package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/your-org/grand-canal-guardian/pkg/errors"
)

type R struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type PageData struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type List struct {
	Items      interface{} `json:"items"`
	Pagination PageData    `json:"pagination"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, R{Code: 0, Message: "ok", Data: data})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, R{Code: 0, Message: "创建成功", Data: data})
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func OKMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, R{Code: 0, Message: message})
}

func OKList(c *gin.Context, items interface{}, page, pageSize int, total int64) {
	totalPages := 0
	if pageSize > 0 {
		totalPages = int(total) / pageSize
		if int(total)%pageSize > 0 {
			totalPages++
		}
	}
	c.JSON(http.StatusOK, R{
		Code:    0,
		Message: "ok",
		Data: List{
			Items: items,
			Pagination: PageData{
				Page: page, PageSize: pageSize, Total: total, TotalPages: totalPages,
			},
		},
	})
}

func Error(c *gin.Context, err *errors.AppError) {
	status := err.Code.HTTPStatus()
	body := R{Code: int(err.Code), Message: err.Message}

	switch errors.Mode {
	case "debug":
		if err.Detail != "" {
			body.Data = map[string]interface{}{"detail": err.Detail}
		}
	default:
		if err.Safe != nil {
			body.Data = map[string]interface{}{"error_type": err.Safe.Type}
		}
	}

	c.JSON(status, body)
	c.Abort()
}

func AbortWithError(c *gin.Context, err *errors.AppError) {
	Error(c, err)
}

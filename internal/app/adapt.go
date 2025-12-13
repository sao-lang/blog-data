package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Adapt(h Handler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ac := &AppContext{
			Context: ctx,
		}
		data, err := h(ac)
		if err != nil {
			if ae, ok := err.(*AppError); ok {
				ctx.JSON(ae.Code, gin.H{
					"code":    ae.Code,
					"message": ae.Message,
					"data":    nil,
				})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code":    http.StatusInternalServerError,
				"message": "internal server error",
				"data":    nil,
			})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"data":    data,
			"message": "success",
		})
	}
}

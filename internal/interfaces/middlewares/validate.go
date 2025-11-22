package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	resp "blog/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// validateParamsError 通用友好提示
func validateParamsError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}

	// 尝试类型断言为 validator.ValidationErrors
	validateErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return false
	}

	// 拼接所有字段的错误信息
	var errMsgs []string
	for _, e := range validateErrors {
		// 使用 Field + Tag + Param 自动生成提示
		msg := fmt.Sprintf("%s failed on '%s' validation", e.Field(), e.Tag())
		if e.Param() != "" {
			msg += fmt.Sprintf(" (param=%s)", e.Param())
		}
		errMsgs = append(errMsgs, msg)
	}

	// 返回 400
	resp.SetCtxResponse(c, nil, http.StatusBadRequest, strings.Join(errMsgs, "; "))
	c.Abort()
	return true
}

// Validate 通用参数验证中间件
func Validate(dto interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 绑定参数
		err := c.ShouldBind(dto)
		if isValidatedError := validateParamsError(c, err); isValidatedError {
			return
		}

		// 其他绑定错误
		if err != nil {
			resp.SetCtxResponse(c, nil, http.StatusBadRequest, err.Error())
			c.Abort()
			return
		}

		// 成功，将 DTO 存入 Context
		resp.SetCtxValidatedData(c, dto)
		c.Next()
	}
}

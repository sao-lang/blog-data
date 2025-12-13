package response

import (
	"blog/internal/common/constants"

	"github.com/gin-gonic/gin"
)

func SetCtxResponse(c *gin.Context, response interface{}, code int, message string) {
	c.Set(constants.RESPONSE_KEY, response)
	c.Set(constants.RESPONSE_CODE_KEY, code)
	c.Set(constants.RESPONSE_MESSAGE_KEY, message)
}

func SetCtxValidatedData(c *gin.Context, dto interface{}) {
	c.Set(constants.VALIDATED_DTO_DATA, dto)
}

func GetCtxResponseStatusCode(c *gin.Context) (int, bool) {
	v, exists := c.Get(constants.RESPONSE_CODE_KEY)
	if !exists {
		return 0, false
	}
	code, ok := v.(int)
	if !ok {
		return 0, false
	}

	return code, true
}

func GetCtxResponseMessage(c *gin.Context) (string, bool) {
	message, exists := c.Get(constants.RESPONSE_MESSAGE_KEY)
	return message.(string), exists
}

func GetCtxResponseData(c *gin.Context) (any, bool) {
	return c.Get(constants.RESPONSE_KEY)
}

func GetCtxValidatedData(c *gin.Context) (any, bool) {
	return c.Get(constants.VALIDATED_DTO_DATA)
}

package middlewares

import (
	"blog/internal/common/constants"
	"blog/internal/config"
	resp "blog/internal/pkg/response"
	"errors"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func verifyToken(tokenString string) (bool, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return false, err
	}
	token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.SecretKey), nil
	})
	if err != nil {
		return false, err
	}
	_, ok := token.Claims.(*jwt.StandardClaims)
	if !ok || !token.Valid {
		return false, errors.New("token verification failed")
	}
	return true, nil
}

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader(constants.TOKEN_KEY)
		if token == "" {
			resp.SetCtxResponse(c, nil, http.StatusUnauthorized, "Token must be not empty")
			return
		}

		isValidatedSuccess, err := verifyToken(token)
		if err != nil || !isValidatedSuccess {
			resp.SetCtxResponse(c, nil, http.StatusUnauthorized, err.Error())
			return
		}
		c.Next()
	}
}

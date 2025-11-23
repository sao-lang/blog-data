package handlers

import (
	"blog/internal/domain/user"
	"blog/internal/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *user.UserService
}

func NewUserHandler(userService *user.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) Register(c *gin.Context) {
	userInfo, _ := response.GetCtxValidatedData(c)
	if user, err := h.userService.Register(userInfo.(*user.CreateUserDTO)); err != nil {
		response.SetCtxResponse(c, user, http.StatusFound, err.Error())
		return
	}
	response.SetCtxResponse(c, nil, http.StatusCreated, "Registered successfully!")
}

func (h *UserHandler) Login(c *gin.Context) {
	userInfo, _ := response.GetCtxValidatedData(c)
	_, accessToken, refreshToken, err := h.userService.Authenticate(userInfo.(*user.CreateUserDTO).UserName, userInfo.(*user.CreateUserDTO).Password)

	if err != nil {
		response.SetCtxResponse(c, nil, http.StatusUnauthorized, err.Error())
		return
	}
	response.SetCtxResponse(c, gin.H{"accessToken": accessToken, "refreshToken": refreshToken}, http.StatusOK, "Login successfully!")
}

func (h *UserHandler) RefreshToken(c *gin.Context) {
	token, _ := response.GetCtxValidatedData(c)
	accessToken, err := h.userService.RefreshToken(token.(*user.RefreshTokenDto).RefreshToken)
	if err != nil {
		response.SetCtxResponse(c, nil, http.StatusForbidden, err.Error())
		return
	}
	response.SetCtxResponse(c, gin.H{"accessToken": accessToken}, http.StatusOK, "Refresh accessToken successfully!")
}

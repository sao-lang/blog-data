package handlers

import (
	"blog/internal/domain/user"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	// 自动注入 Service
	Svc *user.UserService
}

// gnest 会自动将 Body 绑定到 dto，并根据 DTO 里的 binding 标签校验
func (ctrl *UserController) Register(dto *user.CreateUserDTO) interface{} {
	u, err := ctrl.Svc.Register(dto)
	if err != nil {
		return err // 返回 error 会被 gnest 过滤器捕获
	}
	return u
}

func (ctrl *UserController) Login(dto *user.CreateUserDTO) interface{} {
	u, access, refresh, err := ctrl.Svc.Authenticate(dto.UserName, dto.Password)
	if err != nil {
		return err
	}
	return gin.H{
		"user":         u,
		"accessToken":  access,
		"refreshToken": refresh,
	}
}

func (ctrl *UserController) RefreshToken(dto *user.RefreshTokenDto) interface{} {
	newAccess, err := ctrl.Svc.RefreshToken(dto.RefreshToken)
	if err != nil {
		return err
	}
	return gin.H{"accessToken": newAccess}
}

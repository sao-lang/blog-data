package user

import (
	"blog/internal/common/constants"
	"time"
)

type UserDto struct {
	UserName string    `json:"userName" binding:"required"`
	Password string    `json:"password" binding:"required,min=6"`
	Email    string    `json:"email" binding:"email"`
	Phone    string    `json:"phone"`
	FullName string    `json:"fullName"`
	Avatar   string    `json:"avatar"`
	Gender   string    `json:"gender"`
	Birthday time.Time `json:"birthday"`
	Address  string    `json:"address"`
}

type CreateUserDTO struct {
	UserDto
	Role constants.Role `json:"role"`
}

type UpdateUserDTO struct {
	UserDto
	Role   int    `json:"role"`
	Status string `json:"status"`
}

type RefreshTokenDto struct {
	RefreshToken string `json:"refreshToken"`
}

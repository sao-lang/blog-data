package user

import (
	"blog/internal/common/constants"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           string           `gorm:"primaryKey" json:"id"`
	UserName     string           `gorm:"not null" json:"userName"`
	Password     string           `gorm:"not null" json:"password"`
	Email        string           `gorm:"" json:"email"`
	Phone        string           `gorm:"" json:"phone"`
	FullName     string           `gorm:"" json:"fullName"`
	Avatar       string           `gorm:"" json:"avatar"`
	Role         constants.Role   `gorm:"not null" json:"role"`
	Status       constants.Status `gorm:"not null" json:"status"`
	Gender       string           `gorm:"" json:"gender"`
	Birthday     time.Time        `gorm:"" json:"birthday"`
	Address      string           `gorm:"" json:"address"`
	LastLoginAt  time.Time        `gorm:"" json:"lastLoginAt"`
	CreatedAt    time.Time        `gorm:"autoCreateTime" json:"createAt"`
	UpdatedAt    time.Time        `gorm:"autoUpdateTime" json:"updateAt"`
	Salt         string           `gorm:"not null" json:"salt"`
	RefreshToken string           `gorm:"" json:"refreshToken"`
	DeletedAt    gorm.DeletedAt   `gorm:"index" json:"deletedAt"`
}

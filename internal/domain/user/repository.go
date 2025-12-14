package user

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository struct {
	// gnest 会自动注入在 app.Provide 注册过的 *gorm.DB
	DB *gorm.DB
}

func (r *UserRepository) Create(user *User) error {
	user.ID = uuid.NewString()
	return r.DB.Create(&user).Error
}

func (r *UserRepository) FindByUserName(userName string) (*User, error) {
	var user User
	if err := r.DB.Where("user_name = ?", userName).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(userId string, userInfo *User) error {
	var user User
	if err := r.DB.First(&user, userId).Error; err != nil {
		return err
	}
	return r.DB.Model(&user).Updates(&userInfo).Error
}

func (r *UserRepository) Delete(user *User) error {
	return r.DB.Delete(&user).Error
}
